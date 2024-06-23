package db

import (
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Init(connString string) {
	slog.Info("Initializing database...")

	m, err := migrate.New(
		"file://db/migrations",
		connString)

	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			slog.Info("Db Migration Complete", "status", err)
		} else {
			log.Fatal(err)
		}
	}
}
