package order

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

func Test_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		ts, teardown := newTestStorage(t)
		t.Cleanup(teardown)

		// create dummy user
		testUser, err := testCreateUser(t, ts.db)
		if err != nil {
			t.Fatalf("unexpected error when creating dummy user: %v", err)
		}

		books, err := testGetBooks(t, ts.db)
		if err != nil || len(books) < 1 {
			t.Fatalf("unexpected error when getting books: %v", err)
		}
		book := books[0]

		bookPriceFloat, err := strconv.ParseFloat(book.Price, 64)
		if err != nil {
			t.Fatalf("unexpected error when converting price to float64: %v", err)
		}

		validQuantity := int64(1)
		totalBookPrice := fmt.Sprintf("%.2f", float64(validQuantity)*bookPriceFloat)

		testOrder := &models.Order{
			UserID: testUser.ID,
			Total:  totalBookPrice,
		}
		testItems := []*models.OrderItem{
			{
				BookID:   book.ID,
				Price:    book.Price,
				Quantity: validQuantity,
			},
		}

		if err := ts.Create(ctx, testOrder, testItems); err != nil {
			t.Fatalf("Create(_, _, _) expected nil error, got = %v", err)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		notFoundUserID := int64(100000)
		notFoundBookID := int64(100000)

		tests := []struct {
			name    string
			userID  int64
			bookID  int64
			wantErr error
		}{
			{
				name:    "UserIDNotFound",
				userID:  notFoundUserID,
				wantErr: ErrUserIDNotFound,
			},
			{
				name:    "BookIDNotFound",
				bookID:  notFoundBookID,
				wantErr: ErrBookIDNotFound,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ts, teardown := newTestStorage(t)
				t.Cleanup(teardown)

				// create dummy user
				testUser, err := testCreateUser(t, ts.db)
				if err != nil {
					t.Fatalf("unexpected error when creating dummy user: %v", err)
				}

				books, err := testGetBooks(t, ts.db)
				if err != nil || len(books) < 1 {
					t.Fatalf("unexpected error when getting books: %v", err)
				}
				book := books[0]

				bookPriceFloat, err := strconv.ParseFloat(book.Price, 64)
				if err != nil {
					t.Fatalf("unexpected error when converting price to float64: %v", err)
				}

				validQuantity := int64(1)
				totalBookPrice := fmt.Sprintf("%.2f", float64(validQuantity)*bookPriceFloat)

				testUserID := tt.userID
				if testUserID == 0 {
					testUserID = testUser.ID
				}
				testBookID := tt.bookID
				if testBookID == 0 {
					testBookID = book.ID
				}

				testOrder := &models.Order{
					UserID: testUserID,
					Total:  totalBookPrice,
				}
				testItems := []*models.OrderItem{
					{
						BookID:   testBookID,
						Price:    book.Price,
						Quantity: validQuantity,
					},
				}

				err = ts.Create(ctx, testOrder, testItems)
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Create(_, _, _) error, got = %v, want = %v", err, tt.wantErr)
				}
			})
		}
	})
}

func testCreateUser(t *testing.T, db *sqlx.DB) (*models.User, error) {
	t.Helper()

	dummyUser := &models.User{
		Email:    genString(),
		Password: genString(),
	}

	stmt := `INSERT INTO users(%s) VALUES(%s);`

	// fields and values to be operated
	fields := []string{
		"email",
		"password",
	}
	values := []string{
		":email",
		":password",
	}

	res, err := db.NamedExec(
		fmt.Sprintf(stmt, strings.Join(fields, ","), strings.Join(values, ",")),
		dummyUser,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to perform Create operation: %w", err)
	}

	insertedID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to Create user: %v", err)
	}

	createdUser := &models.User{
		ID:       insertedID,
		Email:    dummyUser.Email,
		Password: dummyUser.Password,
	}

	return createdUser, nil
}

func testGetBooks(t *testing.T, db *sqlx.DB) ([]*models.Book, error) {
	query := `
	SELECT id, title, author, price, description
	FROM books
	`

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query from books table: %v", err)
	}
	defer rows.Close()

	// iterate through each row and save it as book model.
	books := []*models.Book{}
	for rows.Next() {
		var book models.Book

		err := rows.StructScan(&book)
		if err != nil {
			return nil, fmt.Errorf("failed when scanning through rows: %v", err)
		}

		books = append(books, &book)
	}

	return books, nil
}

func genString() string {
	return uuid.New().String()
}
