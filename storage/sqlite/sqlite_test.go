package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestNewStorage(t *testing.T) {
	t.Parallel()

	migrationDir := filepath.Join("..", "migrations")

	dbPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected error when getting current path: %v", err)
	}
	pathToDB := filepath.Join(dbPath, "databases")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		testDBName := filepath.Join(pathToDB, uuid.New().String())
		storage, err := NewStorage(testDBName, migrationDir)
		if err != nil {
			t.Fatalf("NewStorage(_), error create new storage: %v", err)
		}

		storage.Teardown()
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		if _, err := NewStorage("", migrationDir); err == nil {
			t.Fatalf("NewStorage(_), expected non nil error: %v", err)
		}
	})
}
