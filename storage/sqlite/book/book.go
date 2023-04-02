package book

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/wilsonangara/simple-online-book-store/storage/models"
)

//go:generate mockgen -source=book.go -destination=mock/book.go -package=mock
type BookStorage interface {
	// GetBooks fetches all books from our storage.
	GetBooks(context.Context) ([]*models.Book, error)
}

type Storage struct {
	db *sqlx.DB
}

// NewStorage creates a wrapper around book storage.
func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

// GetBooks fetches all books from our storage.
func (s *Storage) GetBooks(ctx context.Context) ([]*models.Book, error) {
	query := `
SELECT id, title, author, price, description
FROM books
`

	rows, err := s.db.Queryx(query)
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
