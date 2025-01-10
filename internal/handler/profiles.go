package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
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
