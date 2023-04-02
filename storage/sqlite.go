package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

const driver = "sqlite3"

// Storage provides a wrapper around an SQLite database.
type Storage struct {
	db       *sql.DB
	path     string
	teardown func()
}

func (s *Storage) Database() *sql.DB {
	return s.db
}

// NewStorage creates a connection to our database with the given
// dbName and return a storage with the connected db.
func NewStorage(dbName, migrationDir string) (*Storage, error) {
	dirPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if dbName != "" {
		dbName = dbName + ".db"
	}

	dbPath := filepath.Join(dirPath, "databases", dbName)

	db, err := sql.Open(driver, dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	teardown := func() {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close database connection: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			log.Fatalf("failed to teardown database: %v", err)
		}
	}

	// execute migrations when migrationDir is provided.
	if migrationDir != "" {
		if err := goose.SetDialect(driver); err != nil {
			teardown()
			return nil, fmt.Errorf("failed to set goose dialect: %w", err)
		}
		if err := goose.Up(db, migrationDir); err != nil {
			teardown()
			return nil, fmt.Errorf("failed to run database migrations: %w", err)
		}
	}

	return &Storage{
		db:       db,
		teardown: teardown,
	}, nil
}
