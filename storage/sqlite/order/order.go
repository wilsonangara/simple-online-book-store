package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/wilsonangara/simple-online-book-store/storage/models"
)

var (
	errForeignKeyConstraint = "FOREIGN KEY constraint failed"

	ErrUserIDNotFound = errors.New("user id not found")
	ErrBookIDNotFound = errors.New("book id not found")
)

//go:generate mockgen -source=order.go -destination=mock/order.go -package=mock
type OrderStorage interface {
	// Create adds a new order for a user and allow them to order multiple
	// books.
	Create(context.Context, *models.Order, []*models.OrderItem) error

	// GetOrderHistory fetches all the orders of a user.
	GetOrderHistory(context.Context, int64) ([]*models.OrderHistory, error)
}

type Storage struct {
	db *sqlx.DB
}

// NewStorage creates a wrapper around order storage.
func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

// Create adds a new order for a user and allow them to order multiple
// books.
func (s *Storage) Create(ctx context.Context, order *models.Order, items []*models.OrderItem) error {
	timeNow := time.Now().UTC()

	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	order.CreatedAt = timeNow
	order.UpdatedAt = timeNow

	orderStmt := `INSERT INTO orders (%s) VALUES(%s);`

	orderFields := []string{
		"user_id",
		"total",
	}
	orderValues := []string{
		":user_id",
		":total",
	}

	res, err := tx.NamedExec(
		fmt.Sprintf(orderStmt, strings.Join(orderFields, ","), strings.Join(orderValues, ",")),
		order,
	)
	if err != nil {
		if strings.Contains(err.Error(), errForeignKeyConstraint) {
			return ErrUserIDNotFound
		}
		return fmt.Errorf("failed to insert order operation: %v", err)
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get order id: %v", err)
	}
	order.ID = orderID

	for _, item := range items {
		item.OrderID = orderID
		item.CreatedAt = timeNow
		item.UpdatedAt = timeNow

		itemStmt := `INSERT INTO order_items (%s) VALUES(%s);`

		itemFields := []string{
			"order_id",
			"book_id",
			"price",
			"quantity",
		}
		itemValues := []string{
			":order_id",
			":book_id",
			":price",
			":quantity",
		}

		res, err := tx.NamedExec(
			fmt.Sprintf(itemStmt, strings.Join(itemFields, ","), strings.Join(itemValues, ",")),
			item,
		)
		if err != nil {
			if strings.Contains(err.Error(), errForeignKeyConstraint) {
				return ErrBookIDNotFound
			}
			return fmt.Errorf("failed to insert order item operation: %v", err)
		}

		itemID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get inserted item id: %v", err)
		}
		item.ID = itemID
	}

	tx.Commit()

	return nil
}

// GetOrderHistory fetches all the orders of a user.
func (s *Storage) GetOrderHistory(ctx context.Context, userID int64) ([]*models.OrderHistory, error) {
	query := `
SELECT 
	o.id,
	o.total,
	oi.price,
	oi.quantity,
	b.title,
	b.author,
	b.description
FROM orders o
JOIN order_items oi
	ON o.id = oi.order_id
JOIN books b
	ON oi.book_id = b.id
WHERE user_id = :user_id;
`

	stmt, err := s.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetOrderHistory statement: %w", err)
	}
	defer stmt.Close()

	arg := map[string]interface{}{
		"user_id": userID,
	}

	rows, err := stmt.Queryx(arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query from orders: %v", err)
	}
	defer rows.Close()

	// iterate through each row and save it as order model.
	ordersMap := map[int64]*models.OrderHistory{}
	for rows.Next() {
		// order := &models.Order{}
		var order models.OrderHistoryData

		if err := rows.StructScan(&order); err != nil {
			return nil, fmt.Errorf("failed when scanning through rows: %v", err)
		}

		_, ok := ordersMap[order.ID]
		if !ok {
			ordersMap[order.ID] = &models.OrderHistory{
				ID:    order.ID,
				Total: order.Total,
				Items: []*models.OrderHistoryItem{
					{
						Price:       order.Price,
						Quantity:    order.Quantity,
						Title:       order.Title,
						Author:      order.Author,
						Description: order.Description,
					},
				},
			}
		} else {
			ordersMap[order.ID].Items = append(ordersMap[order.ID].Items, &models.OrderHistoryItem{
				Price:       order.Price,
				Quantity:    order.Quantity,
				Title:       order.Title,
				Author:      order.Author,
				Description: order.Description,
			})
		}
	}

	orders := []*models.OrderHistory{}
	for _, v := range ordersMap {
		orders = append(orders, v)
	}

	return orders, nil
}
