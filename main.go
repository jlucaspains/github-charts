package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jlucaspains/github-charts/db"
	"github.com/jlucaspains/github-charts/handlers"
	"github.com/jlucaspains/github-charts/jobs"
	"github.com/jlucaspains/github-charts/midlewares"
	"github.com/jlucaspains/github-charts/models"
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

	projectConfigs := []models.JobConfigItem{}
	for i := 1; true; i++ {
		rawUrl, ok := os.LookupEnv(fmt.Sprintf("GH_PROJECT_%d", i))
		if !ok {
			break
		}
		config, err := parseProjectConfig(rawUrl)

		if err != nil {
			slog.Warn("Invalid project configuration", "error", err)
			continue
		}

		projectConfigs = append(projectConfigs, config)
	}

	dataPullJob, _ := jobs.NewDataPullJob(jobCron, queries, projectConfigs)
	dataPullJob.Start()

	return dataPullJob.Stop
}

func parseProjectConfig(rawUrl string) (models.JobConfigItem, error) {
	result := models.JobConfigItem{}

	// break string using format key=value separated by spaces
	parts := strings.Split(rawUrl, " ")

	for _, part := range parts {
		keyValue := strings.Split(part, "=")

		if len(keyValue) != 2 {
			continue
		}

		key := keyValue[0]
		value := keyValue[1]

		switch key {
		case "project":
			result.Project = value
		case "org_name":
			result.OrgName = value
		case "repo_owner":
			result.RepoOwner = value
		case "repo_name":
			result.RepoName = value
		case "token":
			result.Token = value
		}
	}

	err := result.Validate()

	return result, err
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

	router.HandleFunc("GET /api/projects", handlers.GetProjects)
	router.HandleFunc("GET /api/projects/{projectId}/burnup", handlers.GetBurnup)
	router.HandleFunc("GET /api/projects/{projectId}/iterations", handlers.GetIterations)
	router.HandleFunc("GET /api/projects/{projectId}/iterations/{iterationId}/burndown", handlers.GetBurndown)
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

	db.Init(dbConnection)

	conn, err := pgxpool.New(ctx, dbConnection)
	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(conn)

	return queries, conn.Close
}
