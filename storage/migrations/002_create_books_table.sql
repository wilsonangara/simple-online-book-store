-- +goose Up
CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        price TEXT NOT NULL,
        description TEXT,
        created_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc')),
        updated_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc'))
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS set_books_updated_at
        AFTER UPDATE ON books
        FOR EACH ROW
        BEGIN
                UPDATE books
                        SET updated_at = DATETIME('now', 'utc')
                        WHERE id = NEW.id;
        END;

INSERT INTO books (title, author, description, price)
        VALUES  ('Atomic Habits', 'James Clear', '', '10.00'),
                ('The Tipping Point', 'Malcolm Gladwell', 'Everything has their own tipping points', '9.80'),
                ('Building a Second brain', 'Tiago Forte', 'A Proven method to organize your digital life and unlock your creative potential', '11.20');
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS set_books_updated_at;
DROP TABLE IF EXISTS books;