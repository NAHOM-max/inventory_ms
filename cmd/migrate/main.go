package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/lib/pq"
)

func main() {
	dir := flag.String("dir", "up", "migration direction: up or down")
	flag.Parse()

	if *dir != "up" && *dir != "down" {
		log.Fatalf("invalid direction %q: must be 'up' or 'down'", *dir)
	}

	//dsn := os.Getenv("DATABASE_URL")
	dsn := "postgres://postgres:N@localhost:5432/ecom_inventory?sslmode=disable"
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	sqlFile := filepath.Join(migrationsDir(), "001_"+*dir+".sql")

	query, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("read %s: %v", sqlFile, err)
	}

	if _, err := db.Exec(string(query)); err != nil {
		log.Fatalf("execute migration: %v", err)
	}

	log.Printf("migration %s applied successfully", *dir)
}

// migrationsDir resolves the migrations/ folder relative to this source file
// so the command works regardless of the working directory it is run from.
func migrationsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("could not resolve source file path")
	}
	// filename = .../cmd/migrate/main.go  →  go up two levels to project root
	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}
