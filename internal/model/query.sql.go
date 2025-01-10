// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package model

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO user (email, password, username) VALUES (?, ?, ?) RETURNING id, username, email, password, bio, image, created_at, updated_at
`

type CreateUserParams struct {
	Email    string
	Password string
	Username string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.Password, arg.Username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Bio,
		&i.Image,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getFollowingCount = `-- name: GetFollowingCount :one
SELECT COUNT(*) FROM following
WHERE user_id = ?
AND follower_id = ?
`

type GetFollowingCountParams struct {
	UserID     int64
	FollowerID int64
}

func (q *Queries) GetFollowingCount(ctx context.Context, arg GetFollowingCountParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, getFollowingCount, arg.UserID, arg.FollowerID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, password, bio, image, created_at, updated_at FROM user WHERE email = ? LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Bio,
		&i.Image,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, username, email, password, bio, image, created_at, updated_at FROM user WHERE id = ? LIMIT 1
`

func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Bio,
		&i.Image,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, email, password, bio, image, created_at, updated_at FROM user WHERE username = ? LIMIT 1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Bio,
		&i.Image,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserByID = `-- name: UpdateUserByID :one
UPDATE user SET
    email = COALESCE(?1, email),
    username = COALESCE(?2, username),
    password = COALESCE(?3, password),
    image = COALESCE(?4, image),
    bio = COALESCE(?5, bio),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?6
RETURNING username, email, bio, image
`

type UpdateUserByIDParams struct {
	Email    *string
	Username *string
	Password *string
	Image    *string
	Bio      *string
	ID       int64
}

type UpdateUserByIDRow struct {
	Username string
	Email    string
	Bio      *string
	Image    *string
}

func (q *Queries) UpdateUserByID(ctx context.Context, arg UpdateUserByIDParams) (UpdateUserByIDRow, error) {
	row := q.db.QueryRowContext(ctx, updateUserByID,
		arg.Email,
		arg.Username,
		arg.Password,
		arg.Image,
		arg.Bio,
		arg.ID,
	)
	var i UpdateUserByIDRow
	err := row.Scan(
		&i.Username,
		&i.Email,
		&i.Bio,
		&i.Image,
	)
	return i, err
}
