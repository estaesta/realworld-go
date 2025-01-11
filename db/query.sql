-- name: CreateUser :one
INSERT INTO user (email, password, username) VALUES (?, ?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM user WHERE email = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM user WHERE id = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM user WHERE username = ? LIMIT 1;

-- name: UpdateUserByID :one
UPDATE user SET
    email = COALESCE(sqlc.narg('email'), email),
    username = COALESCE(sqlc.narg('username'), username),
    password = COALESCE(sqlc.narg('password'), password),
    image = COALESCE(sqlc.narg('image'), image),
    bio = COALESCE(sqlc.narg('bio'), bio),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
RETURNING username, email, bio, image;

-- name: GetFollowingCount :one
SELECT COUNT(*) FROM following
WHERE user_id = ?
AND follower_id = ?;

-- name: FollowByUserUsernameAndFollowerID :exec
INSERT INTO following (user_id, follower_id)
SELECT u.id, ?
FROM user u
WHERE u.username = sqlc.arg('username');
