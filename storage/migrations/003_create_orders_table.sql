-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        total TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        updated_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose StatementBegin
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS orders;
