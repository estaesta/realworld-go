package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

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

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id, ok := claims["user_id"].(float64)
	var user_id_int64 int64
	if ok {
		user_id_int64 = int64(user_id)
	} else {
		user_id_int64 = 0
	}

	comments, err := h.Queries.GetComments(r.Context(), model.GetCommentsParams{
		UserID: user_id_int64,
		Slug:   slug,
	})
	if err != nil && err != sql.ErrNoRows {
		h.serverError(w, err)
		return
	}

	res := map[string]interface{}{
		"comments": []map[string]interface{}{},
	}
	for _, v := range comments {
		res["comments"] = append(res["comments"].([]map[string]interface{}), map[string]interface{}{
			"id":        v.Comment.ID,
			"body":      v.Comment.Body,
			"createdAt": v.Comment.CreatedAt,
			"updatedAt": v.Comment.UpdatedAt,
			"author": map[string]interface{}{
				"username":  v.User.Username,
				"bio":       v.User.Bio,
				"image":     v.User.Image,
				"following": v.IsFollowing > 0,
			},
		})
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		h.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	id := chi.URLParam(r, "id")
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id := int64(claims["user_id"].(float64))

	rows, err := h.Queries.DeleteCommentByIDAndSlug(r.Context(), model.DeleteCommentByIDAndSlugParams{
		ID:     idInt64,
		Slug:   slug,
		UserID: user_id,
	})
	if rows == 0 {
		h.clientError(w, http.StatusBadRequest)
		return
	}
	if err != nil {
		h.serverError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) FavoriteArticle(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id := int64(claims["user_id"].(float64))

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	rows, err := qtx.FavoritArticle(r.Context(), model.FavoritArticleParams{
		UserID: user_id,
		Slug:   slug,
	})
	if rows == 0 {
		h.clientError(w, http.StatusNotFound)
		return
	}
	if err != nil {
		h.serverError(w, err)
		return
	}

	article, err := qtx.GetArticleBySlug(r.Context(), model.GetArticleBySlugParams{
		UserID: user_id,
		Slug:   slug,
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
		"article": map[string]interface{}{
			"slug":        article.Article.Slug,
			"title":       article.Article.Title,
			"description": article.Article.Description,
			"body":        article.Article.Body,
			// "tagList":        strings.Split(article.Tags.(string), ","),
			"tagList": func() []string {
				tags := strings.Split(article.Tags.(string), ",")
				sort.Strings(tags)
				return tags
			}(),
			"createdAt":      article.Article.CreatedAt,
			"updatedAt":      article.Article.UpdatedAt,
			"favorited":      article.Favorited > 0,
			"favoritesCount": article.FavoritesCount,
			"author": map[string]interface{}{
				"username":  article.User.Username,
				"bio":       article.User.Bio,
				"image":     article.User.Image,
				"following": article.IsFollowing > 0,
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

func (h *Handler) UnfavoriteArticle(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id := int64(claims["user_id"].(float64))

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	rows, err := qtx.UnfavoriteArticle(r.Context(), model.UnfavoriteArticleParams{
		UserID: user_id,
		Slug:   slug,
	})
	if rows == 0 {
		h.clientError(w, http.StatusNotFound)
		return
	}
	if err != nil {
		h.serverError(w, err)
		return
	}

	article, err := qtx.GetArticleBySlug(r.Context(), model.GetArticleBySlugParams{
		UserID: user_id,
		Slug:   slug,
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
		"article": map[string]interface{}{
			"slug":        article.Article.Slug,
			"title":       article.Article.Title,
			"description": article.Article.Description,
			"body":        article.Article.Body,
			// "tagList":        strings.Split(article.Tags.(string), ","),
			"tagList": func() []string {
				tags := strings.Split(article.Tags.(string), ",")
				sort.Strings(tags)
				return tags
			}(),
			"createdAt":      article.Article.CreatedAt,
			"updatedAt":      article.Article.UpdatedAt,
			"favorited":      article.Favorited > 0,
			"favoritesCount": article.FavoritesCount,
			"author": map[string]interface{}{
				"username":  article.User.Username,
				"bio":       article.User.Bio,
				"image":     article.User.Image,
				"following": article.IsFollowing > 0,
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

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.Queries.GetAllTagsName(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}

	sort.Strings(tags)

	res := map[string]interface{}{
		"tags": tags,
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		h.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}
