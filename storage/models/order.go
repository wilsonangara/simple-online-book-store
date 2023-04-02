package models

import "time"

type Order struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Total     string    `db:"total"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type OrderItem struct {
	ID        int64     `db:"id"`
	OrderID   int64     `db:"order_id"`
	BookID    int64     `db:"book_id"`
	Price     string    `db:"price"`
	Quantity  int64     `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type OrderHistoryData struct {
	ID          int64  `db:"id"`
	Total       string `db:"total"`
	Price       string `db:"price"`
	Quantity    int64  `db:"quantity"`
	Title       string `db:"title"`
	Author      string `db:"author"`
	Description string `db:"description"`
}

type OrderHistoryItem struct {
	Price       string
	Quantity    int64
	Title       string
	Author      string
	Description string
}

type OrderHistory struct {
	ID    int64
	Total string
	Items []*OrderHistoryItem
}
