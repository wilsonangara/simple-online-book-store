-- +goose Up
CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        updated_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc'))
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS set_user_updated_at
        AFTER UPDATE ON users
        FOR EACH ROW
        BEGIN
                UPDATE users
                        SET updated_at = DATETIME('now', 'utc')
                        WHERE id = NEW.id;
        END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS set_user_updated_at;
DROP TABLE IF EXISTS users;
