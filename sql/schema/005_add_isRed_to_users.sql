-- +goose Up
ALTER TABLE users add COLUMN is_chirpy_red boolean not null DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN is_chirpy_red;

