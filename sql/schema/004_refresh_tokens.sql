-- +goose Up
CREATE TABLE refresh_tokens(token TEXT not null, created_at TIMESTAMP NOT NULL default CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL default CURRENT_TIMESTAMP, user_id UUID NOT NULL,  expires_at TIMESTAMP NOT NULL,  revoked_at TIMESTAMP, CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE cascade);

-- +goose Down
DROP TABLE refresh_tokens;
