package storage

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestNewStorage(t *testing.T) {
	t.Parallel()

	migrationDir := filepath.Join(".", "migrations")

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		testDBName := uuid.New()
		storage, err := NewStorage(testDBName.String(), migrationDir)
		if err != nil {
			t.Fatalf("NewStorage(_), error create new storage: %v", err)
		}

		storage.teardown()
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		if _, err := NewStorage("", migrationDir); err == nil {
			t.Fatalf("NewStorage(_), expected non nil error: %v", err)
		}
	})
}
