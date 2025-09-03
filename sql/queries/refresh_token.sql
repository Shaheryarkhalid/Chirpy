-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, user_id, expires_at) values($1, $2, $3) returning *; 

-- name: GetRefreshToken :one
SELECT * from refresh_tokens where token = $1 LIMIT 1;

-- name: GetUserFromRefreshToken :one
SELECT refresh_tokens.token,users.* from refresh_tokens join users on refresh_tokens.user_id = users.id where refresh_tokens.token = $1 LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = $1, updated_at = $2 where token = $3;
