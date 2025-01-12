package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/jwtauth/v5"
)

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) FeedArticles(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Article struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Body        string `json:"body"`
			TagList     []string
		} `json:"article"`
	}
	var req = request{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	// Request validation
	if req.Article.Title == "" || req.Article.Description == "" || req.Article.Body == "" {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id := int64(claims["user_id"].(float64))

	// create slug
	slug := strings.TrimSpace(req.Article.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = regexp.MustCompile(`[^a-zA-Z0-9\-]+`).ReplaceAllString(slug, "")
	slug = strings.ToLower(slug)
	slug = fmt.Sprintf("%s-%s", slug, time.Now().Format("20060102150405"))

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Create article
	articleID, err := qtx.CreateArticle(r.Context(), model.CreateArticleParams{
		Slug:        slug,
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		AuthorID:    user_id,
	})
	if err != nil {
		h.serverError(w, err)
		return
	}

	tagListInterface := make([]interface{}, len(req.Article.TagList))
	for i, v := range req.Article.TagList {
		tagListInterface[i] = v
	}
	// Get tags
	tags, err := qtx.GetTags(r.Context(), req.Article.TagList)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create array of tagIDs from tags[].ID
	tagIDs := make([]int64, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}

	// Check if we need to create additional tags
	if len(tags) < len(req.Article.TagList) {
		tagNames := make([]string, len(tags))
		for i, v := range tags {
			tagNames[i] = v.Name
		}
		for _, v := range req.Article.TagList {
			if slices.Contains(tagNames, v) {
				continue
			}
			additionalTagIDs, err := qtx.CreateTag(r.Context(), v)
			if err != nil {
				h.serverError(w, err)
				return
			}
			tagIDs = append(tagIDs, additionalTagIDs.ID)
		}
	}

	// Create article tags
	for _, v := range tagIDs {
		err = qtx.CreateArticleTag(r.Context(), model.CreateArticleTagParams{
			ArticleID: articleID.ID,
			TagID:     v,
		})
		if err != nil {
			h.serverError(w, err)
			return
		}
	}

	// Get author aka current user
	user, err := qtx.GetUserByID(r.Context(), user_id)
	if err != nil {
		h.serverError(w, err)
		return
	}

	// commit
	err = tx.Commit()
	if err != nil {
		h.serverError(w, err)
		return
	}

	res := map[string]interface{}{
		"article": map[string]interface{}{
			"slug":           slug,
			"title":          req.Article.Title,
			"description":    req.Article.Description,
			"body":           req.Article.Body,
			"tagList":        req.Article.TagList,
			"createdAt":      articleID.CreatedAt,
			"updatedAt":      articleID.CreatedAt,
			"favorited":      false,
			"favoritesCount": 0,
			"author": map[string]interface{}{
				"username":  user.Username,
				"bio":       user.Bio,
				"image":     user.Image,
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
