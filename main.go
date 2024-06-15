package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jlucaspains/github-charts/db"
	"github.com/jlucaspains/github-charts/handlers"
	"github.com/jlucaspains/github-charts/jobs"
	"github.com/jlucaspains/github-charts/midlewares"
	"github.com/joho/godotenv"
)

func loadEnv() {
	// outside of local environment, variables should be
	// OS environment variables
	env := os.Getenv("ENV")
	if err := godotenv.Load(); err != nil && env == "" {
		log.Fatal(fmt.Printf("Error loading .env file: %s", err))
	}
}

func main() {
	loadEnv()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()
	queries, dispose := initDB(ctx)
	defer dispose()

	dataPullJobDispose := startDataPullJob(queries)
	defer dataPullJobDispose()

	webDispose := startWebServer(queries)
	defer webDispose(ctx)

	<-done

	slog.Info("Stopping jobs...")
	slog.Info("Stopping web server...")
}

func startDataPullJob(queries *db.Queries) func() {
	jobCron := os.Getenv("DATA_PULL_JOB_CRON")
	if jobCron == "" {
		log.Fatalf("must set DATA_PULL_JOB_CRON=<CRON>")
	}

	key := os.Getenv("GH_TOKEN")
	if key == "" {
		log.Fatalf("must set GITHUB_TOKEN=<github token>")
	}

	orgName := os.Getenv("GH_ORG_NAME")
	if orgName == "" {
		log.Fatalf("must set ORG_NAME=<organization name>")
	}

	projectIdConfig := os.Getenv("PROJECT_ID")
	projectId, err := strconv.Atoi(projectIdConfig)

	if err != nil || projectId <= 0 {
		log.Fatalf("must set PROJECT_ID=<project id>")
	}

	dataPullJob := &jobs.DataPullJob{}
	dataPullJob.Init(jobCron, queries, int(projectId), key, orgName)

	dataPullJob.Start()

	return dataPullJob.Stop
}

func getAllowedOrigins() string {
	allowedOrigin, ok := os.LookupEnv("ALLOWED_ORIGIN")
	if !ok {
		allowedOrigin = "http://localhost:5173"
	}

	return allowedOrigin
}

func startWebServer(queries *db.Queries) func(ctx context.Context) error {
	handlers := &handlers.Handlers{Queries: queries, CORSOrigins: getAllowedOrigins()}

	router := http.NewServeMux()

	router.HandleFunc("GET /api/iterations", handlers.GetIterations)
	router.HandleFunc("GET /api/iterations/{id}/burndown", handlers.GetBurndown)
	router.HandleFunc("GET /api/projects", handlers.GetProjects)
	router.HandleFunc("GET /api/projects/{id}/burnup", handlers.GetBurnup)
	router.HandleFunc("GET /health", handlers.HealthCheck)

	if handlers.CORSOrigins != "" {
		router.HandleFunc("OPTIONS /api/", handlers.CORS)
	}

	router.Handle("/", http.FileServer(http.Dir("./public/")))

	logRouter := midlewares.NewLogger(router)

	hostPort, ok := os.LookupEnv("WEB_HOST_PORT")
	if !ok {
		hostPort = ":8000"
	}

	certFile, useTls := os.LookupEnv("TLS_CERT_FILE")

	certKeyFile, ok := os.LookupEnv("TLS_CERT_KEY_FILE")
	useTls = useTls && ok

	slog.Info("Starting TLS server", "port", hostPort, "usetls", useTls)

	srv := &http.Server{
		Addr: hostPort,
	}

	srv.Handler = logRouter

	go func() {
		var err error = nil
		if useTls {
			err = srv.ListenAndServeTLS(certFile, certKeyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	slog.Info("Web Server Started")
	return srv.Shutdown
}

func initDB(ctx context.Context) (*db.Queries, func()) {
	dbConnection := os.Getenv("DB_CONNECTION")

	if dbConnection == "" {
		log.Fatal("must set DB_CONNECTION=<connection string>")
	}

	conn, err := pgxpool.New(ctx, dbConnection)
	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(conn)

	return queries, conn.Close
}
