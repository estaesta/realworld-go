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

-- name: IsUserFollowByUserID :one
SELECT COUNT(*) FROM following
WHERE user_id = ?
AND follower_id = ?;

-- name: FollowByUserUsernameAndFollowerID :exec
INSERT INTO following (user_id, follower_id)
SELECT u.id, ?
FROM user u
WHERE u.username = sqlc.arg('username');

-- name: UnfollowByUserIDAndFollowerID :exec
DELETE FROM following
WHERE user_id = sqlc.arg('user_id')
AND follower_id = sqlc.arg('follower_id');

-- name: CreateArticle :one
INSERT INTO article (slug, title, description, body, author_id) 
VALUES (?, ?, ?, ?, ?) 
RETURNING id, created_at;

-- name: CreateTag :one
INSERT OR IGNORE INTO tag (name) VALUES (?)
    RETURNING *;

-- name: GetTags :many
SELECT * FROM tag WHERE name IN(sqlc.slice('tags'));

-- name: CreateArticleTag :exec
INSERT OR IGNORE INTO article_tag (article_id, tag_id)
SELECT ?, ?;

-- name: GetArticleBySlug :one
SELECT sqlc.embed(article), sqlc.embed(user), GROUP_CONCAT(tag.name) AS tag
FROM article 
JOIN user ON article.author_id = user.id
JOIN article_tag ON article.id = article_tag.article_id
JOIN tag ON article_tag.tag_id = tag.id
WHERE slug = ?;

-- name: IsFavoriteByUserIDAndArticleID :one
SELECT COUNT(*) FROM favorite
WHERE user_id = ?
AND article_id = ?;
