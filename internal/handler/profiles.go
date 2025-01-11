package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/mattn/go-sqlite3"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.Queries.GetUserByUsername(r.Context(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			h.clientError(w, http.StatusNotFound)
			return
		}
		h.serverError(w, err)
		return
	}
	isFollowing := false

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id, ok := claims["user_id"].(float64)
	if ok {
		following, err := h.Queries.GetFollowingCount(r.Context(),
			model.GetFollowingCountParams{
				UserID:     user.ID,
				FollowerID: int64(user_id),
			})
		if err != nil {
			h.serverError(w, err)
			return
		}
		isFollowing = following > 0
	}

	res := map[string]interface{}{
		"profile": map[string]interface{}{
			"username":  user.Username,
			"bio":       user.Bio,
			"image":     user.Image,
			"following": isFollowing,
		},
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		h.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}

func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.clientError(w, http.StatusUnauthorized)
		return
	}
	currentUserId := int64(claims["user_id"].(float64))

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	user, err := qtx.GetUserByUsername(r.Context(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			h.clientError(w, http.StatusNotFound)
			return
		}
		h.serverError(w, err)
		return
	}
	err = qtx.FollowByUserUsernameAndFollowerID(r.Context(), model.FollowByUserUsernameAndFollowerIDParams{
		FollowerID: currentUserId,
		Username:   username,
	})
	var sqlite *sqlite3.Error
	if errors.As(err, &sqlite) {
		if sqlite.Code == sqlite3.ErrConstraint {
			h.clientError(w, http.StatusBadRequest)
			return
		}
		h.serverError(w, err)
		return
	}
	err = tx.Commit()
	if err != nil {
		h.serverError(w, err)
		return
	}

	res := map[string]interface{}{
		"profile": map[string]interface{}{
			"username":  user.Username,
			"bio":       user.Bio,
			"image":     user.Image,
			"following": true,
		},
	}

	resJson, err := json.Marshal(res)
	if err != nil {
		h.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}

func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.clientError(w, http.StatusUnauthorized)
		return
	}
	currentUserId := int64(claims["user_id"].(float64))

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	user, err := qtx.GetUserByUsername(r.Context(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			h.clientError(w, http.StatusNotFound)
			return
		}
		h.serverError(w, err)
		return
	}

	err = qtx.UnfollowByUserIDAndFollowerID(r.Context(), model.UnfollowByUserIDAndFollowerIDParams{
		UserID:     user.ID,
		FollowerID: currentUserId,
	})
	if err != nil {
		h.serverError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		h.serverError(w, err)
		return
	}

	res := map[string]interface{}{
		"profile": map[string]interface{}{
			"username":  user.Username,
			"bio":       user.Bio,
			"image":     user.Image,
			"following": false,
		},
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		h.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}
