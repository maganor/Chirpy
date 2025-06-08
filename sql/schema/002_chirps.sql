-- +goose Up
CREATE TABLE chirps (
    id uuid,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    body TEXT NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT fk_chirps_users
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;