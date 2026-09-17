package main

import (
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("usage: migrate<up|down>")
	}

	m, err := migrate.New(
		"file://migrations",
		"sqlite://data/gadget_store.db",
	)

	if err != nil {
		log.Fatalf("migration.new:%v", err)
	}

	defer m.Close()

	switch os.Args[1] {
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force <version>")
		}

		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}

		if err := m.Force(version); err != nil {
			log.Fatal(err)
		}

	log.Printf("migration forced to version %d", version)
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("migration up applied")

	case "down":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("migration down applied")

	default:
		log.Fatalf("unknown command: %s", os.Args[1])

	}

}
