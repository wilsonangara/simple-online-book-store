package book

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
)

func newTestStorage(tb testing.TB) (*Storage, func()) {
	dir, err := os.Getwd()
	if err != nil {
		tb.Fatalf("unexpected error when getting working directory: %v", err)
	}

	testDB := filepath.Join(dir, genString())
	pathToMigrationsDir := filepath.Join("..", "..", "migrations")

	ts, err := sqlite.NewStorage(testDB, pathToMigrationsDir)
	if err != nil {
		tb.Fatalf("failed to create new test storage: %v", err)
	}

	return &Storage{db: ts.Database()}, ts.Teardown
}

func Test_GetBooks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ts, teardown := newTestStorage(t)
	t.Cleanup(teardown)

	_, err := ts.GetBooks(ctx)
	if err != nil {
		t.Fatalf("GetBooks(_) expected nil error, got = %v", err)
	}
}

func genString() string {
	return uuid.New().String()
}
