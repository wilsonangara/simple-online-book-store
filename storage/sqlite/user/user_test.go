package user

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/wilsonangara/simple-online-book-store/storage/models"
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

func Test_GetUserByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ts, teardown := newTestStorage(t)
	t.Cleanup(teardown)

	testEmail := genString()
	testPassword := genString()

	// create dummy user
	createdUser, err := ts.Create(ctx, &models.User{
		Email:    testEmail,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("unexpected error when creating dummy user: %v", err)
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		user, err := ts.GetUserByID(ctx, int(createdUser.ID))
		if err != nil {
			t.Fatalf("GetUserByID(_, _) expected nil error, got = %v", err)
		}

		// check user email
		if user.Email != testEmail {
			t.Fatalf("GetUserByID(_, _) error, got = %v, want = %v",
				createdUser.Email, testEmail,
			)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			id   int
			err  error
		}{
			{
				name: "UserNotFound",
				id:   100000000000,
				err:  sqlite.ErrNotFound,
			},
			{
				name: "InvalidUserID",
				id:   -1,
				err:  ErrInvalidUserID,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				// ignoring the user response as we are only checking the error.
				_, err := ts.GetUserByID(ctx, 100000000)
				if !errors.Is(err, sqlite.ErrNotFound) {
					t.Fatalf("GetUserByID(_, _) error, got = %v, want = %v", err, sqlite.ErrNotFound)
				}
			})
		}
	})
}

func Test_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ts, teardown := newTestStorage(t)
	t.Cleanup(teardown)

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		validUser := &models.User{
			Email:    genString(),
			Password: genString(),
		}

		user, err := ts.Create(ctx, validUser)
		if err != nil {
			t.Fatalf("Create(_, _) expected nil error, got = %v", err)
		}

		// compare user's email
		if user.Email != validUser.Email {
			t.Fatalf("Create(_, _) error, got = %v, want = %v", user.Email, validUser.Email)
		}
	})

	t.Run("Failed_DuplicateEmail", func(t *testing.T) {
		t.Parallel()

		testEmail := genString()
		testPassword := genString()

		// dummyUser will be created in advance to check duplicate email.
		_, err := ts.Create(ctx, &models.User{
			Email:    testEmail,
			Password: testPassword,
		})
		if err != nil {
			t.Fatalf("unexpected error when creating dummy user: %v", err)
		}

		_, err = ts.Create(ctx, &models.User{
			Email:    testEmail,
			Password: testPassword,
		})
		if !errors.Is(err, ErrEmailAlreadyExist) {
			t.Fatalf("Create(_, _) error, got = %v, want = %v", err, ErrEmailAlreadyExist)
		}
	})
}

func genString() string {
	return uuid.New().String()
}
