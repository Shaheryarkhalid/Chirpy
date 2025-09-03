-- +goose UP
ALTER TABLE users ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down 
Alter TABLE users drop COLUMN hashed_password;
