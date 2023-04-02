-- +goose Up
CREATE TABLE IF NOT EXISTS order_items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id INTEGER NOT NULL,
        book_id INTEGER NOT NULL,
        price TEXT NOT NULL,
        quantity INTEGER NOT NULL,
        created_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        updated_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        FOREIGN KEY (order_id) REFERENCES orders(id),
        FOREIGN KEY (book_id) REFERENCES books(id)
)

-- +goose StatementBegin
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS order_items;
