package sqlite

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

const driver = "sqlite3"

var errDbNameIsRequired = errors.New("db name is required")

func init() {
	goose.SetDialect(driver)
}

// Storage provides a wrapper around an SQLite database.
type Storage struct {
	db       *sqlx.DB
	path     string
	Teardown func()
}

func (s *Storage) Database() *sqlx.DB {
	return s.db
}

// NewStorage creates a connection to our database with the given
// dbName and return a storage with the connected db.
func NewStorage(dbName, migrationDir string) (*Storage, error) {
	if dbName == "" {
		return nil, errDbNameIsRequired
	}
	dbName = dbName + ".db"

	db, err := sqlx.Connect(driver, dbName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.DB.Exec(`PRAGMA foreign_keys = ON;`)

	teardown := func() {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close database connection: %v", err)
		}
		if err := os.Remove(dbName); err != nil {
			log.Fatalf("failed to teardown database: %v", err)
		}
	}

	// execute migrations when migrationDir is provided.
	if migrationDir != "" {
		if err := goose.Up(db.DB, migrationDir); err != nil {
			teardown()
			return nil, fmt.Errorf("failed to run database migrations: %w", err)
		}
	}

	return &Storage{
		db:       db,
		Teardown: teardown,
	}, nil
}
