package handler

import (
	"encoding/json"
	"net/http"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id := int64(claims["user_id"].(float64))

	type request struct {
		Comment struct {
			Body string `json:"body"`
		} `json:"comment"`
	}
	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	comment, err := qtx.AddComment(r.Context(), model.AddCommentParams{
		Body:     req.Comment.Body,
		AuthorID: user_id,
		Slug:     slug,
	})
	if err != nil {
		h.serverError(w, err)
		return
	}

	author, err := qtx.GetUserByID(r.Context(), user_id)
	if err != nil {
		h.serverError(w, err)
		return
	}

	tx.Commit()

	res := map[string]interface{}{
		"comment": map[string]interface{}{
			"id":        comment.ID,
			"body":      comment.Body,
			"createdAt": comment.CreatedAt,
			"updatedAt": comment.UpdatedAt,
			"author": map[string]interface{}{
				"username":  author.Username,
				"bio":       author.Bio,
				"image":     author.Image,
				"following": false,
			},
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
