-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password) values( gen_random_uuid() , Now(), Now(), $1, $2) returning *;

-- name: GetUser :one
SELECT * from users where id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * from users where email = $1 LIMIT 1;
 
-- name: UpdateUser :one
UPDATE users set email = $1, hashed_password = $2 where id = $3 returning id, email, created_at, updated_at, is_chirpy_red; 

-- name: DeleteAllUsers :exec
DELETE from users;

-- name: UpgradeUserToRed :one
UPDATE users set is_chirpy_red = true where id = $1 returning *;
