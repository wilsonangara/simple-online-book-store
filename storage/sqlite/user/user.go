package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/wilsonangara/simple-online-book-store/storage/models"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
)

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrEmailAlreadyExist = errors.New("email already exist")
)

//go:generate mockgen -source=user.go -destination=mock/user.go -package=mock
type UserStorage interface {
	// GetUserByID fetches the user in our storage.
	GetUserByID(context.Context, int64) (*models.User, error)

	// Create adds a new user to our storage.
	Create(context.Context, *models.User) (*models.User, error)
}

type Storage struct {
	db *sqlx.DB
}

// NewStorage creates a wrapper around user storage.
func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
	}
}

// GetUserByID fetches the user in our database
func (s *Storage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if id < 1 {
		return nil, ErrInvalidUserID
	}

	query := `
SELECT id, email, password
FROM users
WHERE id = :id
`

	stmt, err := s.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetUserByID statement: %w", err)
	}
	defer stmt.Close()

	var user models.User
	arg := map[string]interface{}{
		"id": id,
	}
	if err := stmt.Get(&user, arg); err != nil {
		if err == sql.ErrNoRows {
			return nil, sqlite.ErrNotFound
		}
		return nil, fmt.Errorf("failed to perform GetUserByID storage operation: %w", err)
	}
	return &user, nil
}

// Create adds a new user to our storage.
func (s *Storage) Create(ctx context.Context, user *models.User) (*models.User, error) {
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

	res, err := s.db.NamedExec(
		fmt.Sprintf(stmt, strings.Join(fields, ","), strings.Join(values, ",")),
		user,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, ErrEmailAlreadyExist
		}
		return nil, fmt.Errorf("failed to perform Create operation: %w", err)
	}

	insertedID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to Create user: %v", err)
	}

	createdUser := &models.User{
		ID:       insertedID,
		Email:    user.Email,
		Password: user.Password,
	}

	return createdUser, nil
}
