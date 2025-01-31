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
RETURNING id, created_at, updated_at;

-- name: CreateTag :one
INSERT OR IGNORE INTO tag (name) VALUES (?)
    RETURNING *;

-- name: GetTags :many
SELECT * FROM tag WHERE name IN(sqlc.slice('tags'));

-- name: GetAllTagsName :many
SELECT name FROM tag;

-- name: CreateArticleTag :exec
INSERT OR IGNORE INTO article_tag (article_id, tag_id)
SELECT ?, ?;

-- name: GetArticleBySlug :one
SELECT sqlc.embed(article), 
    sqlc.embed(user), 
    IFNULL(GROUP_CONCAT(tag.name), '') AS tags,
    CASE
        WHEN EXISTS(
        SELECT 1 FROM favorite WHERE article_id = article.id 
            AND favorite.user_id = sqlc.arg('user_id')
        ) THEN 1
        ELSE 0
    END AS favorited,
    (SELECT COUNT(*) FROM favorite
    WHERE article_id = article.id) AS favorites_count,
    CASE
        WHEN following.user_id IS NOT NULL THEN 1
        ELSE 0
    END AS is_following
FROM article 
JOIN user ON article.author_id = user.id
LEFT JOIN article_tag ON article.id = article_tag.article_id
LEFT JOIN tag ON article_tag.tag_id = tag.id
LEFT JOIN favorite ON article.id = favorite.article_id
LEFT JOIN following ON article.author_id = following.user_id AND following.follower_id = sqlc.arg('user_id')
WHERE slug = sqlc.arg('slug')
GROUP BY article.id;

-- name: IsFavoriteByUserIDAndArticleID :one
SELECT COUNT(*) FROM favorite
WHERE user_id = ?
AND article_id = ?;

-- name: GetArticleAuthorBySlug :one
SELECT author_id FROM article
WHERE slug = ?;

-- name: GetArticlesList :many
SELECT 
    article.id,
    article.slug,
    article.title,
    article.description,
    IFNULL(GROUP_CONCAT(tag.name), '') AS tags,
    CASE
        WHEN EXISTS(
        SELECT 1 FROM favorite WHERE article_id = article.id 
            AND favorite.user_id = sqlc.arg('user_id')
        ) THEN 1
        ELSE 0
    END AS favorited,
    (SELECT COUNT(*) FROM favorite
    WHERE article_id = article.id) AS favorites_count,
    article.created_at,
    article.updated_at,
    sqlc.embed(user),
    CASE
        WHEN following.user_id IS NOT NULL THEN 1
        ELSE 0
    END AS is_following
FROM article
LEFT JOIN user ON article.author_id = user.id
LEFT JOIN article_tag ON article.id = article_tag.article_id
LEFT JOIN tag ON article_tag.tag_id = tag.id
LEFT JOIN favorite ON article.id = favorite.article_id
LEFT JOIN following ON article.author_id = following.user_id AND following.follower_id = sqlc.arg('user_id')
WHERE (user.username = sqlc.arg(author) or sqlc.arg(author) = '')
    AND (article.id IN (
        SELECT article_id FROM article_tag
        JOIN tag ON article_tag.tag_id = tag.id
        WHERE tag.name = sqlc.arg(tag)
        ) OR sqlc.arg(tag) = '')
    AND (favorite.user_id = sqlc.arg(favorited) or sqlc.arg(favorited) = 0)
GROUP BY article.id
ORDER BY article.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetArticlesFeed :many
SELECT 
    article.id,
    article.slug,
    article.title,
    article.description,
    IFNULL(GROUP_CONCAT(tag.name), '') AS tags,
    CASE
        WHEN EXISTS(
        SELECT 1 FROM favorite WHERE article_id = article.id 
            AND favorite.user_id = sqlc.arg('user_id')
        ) THEN 1
        ELSE 0
    END AS favorited,
    (SELECT COUNT(*) FROM favorite
    WHERE article_id = article.id) AS favorites_count,
    article.created_at,
    article.updated_at,
    sqlc.embed(user)
FROM article
LEFT JOIN user ON article.author_id = user.id
LEFT JOIN article_tag ON article.id = article_tag.article_id
LEFT JOIN tag ON article_tag.tag_id = tag.id
LEFT JOIN favorite ON article.id = favorite.article_id
JOIN following ON article.author_id = following.user_id AND following.follower_id = sqlc.arg('user_id')
GROUP BY article.id
ORDER BY article.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdateArticle :one
UPDATE article SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    body = COALESCE(sqlc.narg('body'), body),
    updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
WHERE slug = sqlc.arg('slug')
RETURNING slug;

-- name: DeleteArticleBySlug :exec
DELETE FROM article WHERE slug = sqlc.arg('slug');

-- name: AddComment :one
INSERT INTO comment (body, author_id, article_id)
VALUES (?, ?, (SELECT id FROM article WHERE slug = @slug))
RETURNING *;

-- name: GetComments :many
SELECT sqlc.embed(comment),
    sqlc.embed(user),
    CASE
        WHEN following.user_id IS NOT NULL THEN 1
        ELSE 0
    END AS is_following
FROM comment 
JOIN user ON comment.author_id = user.id
LEFT JOIN following ON comment.author_id = following.user_id AND following.follower_id = sqlc.arg('user_id')
WHERE article_id = (SELECT id FROM article WHERE slug = sqlc.arg('slug'))
ORDER BY comment.created_at DESC;

-- name: DeleteCommentByIDAndSlug :execrows
DELETE FROM comment 
WHERE comment.id = sqlc.arg('id') 
    AND article_id = (SELECT id FROM article WHERE slug = sqlc.arg('slug'))
    AND comment.author_id = sqlc.arg('user_id');

-- name: FavoritArticle :execrows
INSERT INTO favorite (user_id, article_id)
VALUES (sqlc.arg('user_id'), (SELECT id FROM article WHERE slug = sqlc.arg('slug')));

-- name: UnfavoriteArticle :execrows
DELETE FROM favorite
WHERE user_id = sqlc.arg('user_id')
    AND article_id = (SELECT id FROM article WHERE slug = sqlc.arg('slug'));
