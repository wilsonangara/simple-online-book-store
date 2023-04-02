package models

import "time"

type Book struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Author      string    `db:"author"`
	Price       string    `db:"price"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
