-- +goose Up
CREATE TABLE users (
	id uuid UNIQUE NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	email TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;
