-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id TEXT NOT NULL,
        quantity INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        book_id INTEGER NOT NULL,
        created_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        updated_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        FOREIGN KEY (user_id) REFERENCES users(id),
        FOREIGN KEY (book_id) REFERENCES books(id)
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS set_orders_updated_at
        AFTER UPDATE ON orders
        FOR EACH ROW
        BEGIN
                UPDATE orders
                        SET updated_at = DATETIME('now', 'utc')
                        WHERE id = NEW.id;
        END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS set_orders_updated_at;
DROP TABLE IF EXISTS orders;
