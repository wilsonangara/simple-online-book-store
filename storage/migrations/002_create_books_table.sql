-- +goose Up
CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        price TEXT NOT NULL,
        description TEXT,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
INSERT INTO books (title, author, description, price)
        VALUES  ('Atomic Habits', 'James Clear', '', '10.00'),
                ('The Tipping Point', 'Malcolm Gladwell', 'Everything has their own tipping points', '9.80'),
                ('Building a Second brain', 'Tiago Forte', 'A Proven method to organize your digital life and unlock your creative potential', '11.20');
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS books;
