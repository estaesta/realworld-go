package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	author := r.URL.Query().Get("author")
	favorited := r.URL.Query().Get("favorited")

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "20"
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	offset := r.URL.Query().Get("offset")
	if offset == "" {
		offset = "0"
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id, ok := claims["user_id"].(float64)
	var user_id_int64 int64
	if ok {
		user_id_int64 = int64(user_id)
	} else {
		user_id_int64 = 0
	}

	var favoritedParam int64
	if favorited != "" {
		favoritedUserID, err := h.Queries.GetUserByUsername(r.Context(), favorited)
		if err != nil && err != sql.ErrNoRows {
			h.serverError(w, err)
			return
		}
		if err == sql.ErrNoRows {
			//if not found, set to -1 so that the query return empty
			favoritedParam = -1
		}
		if err == nil {
			favoritedParam = favoritedUserID.ID
		}
	}

	articles, err := h.Queries.GetArticlesList(r.Context(), model.GetArticlesListParams{
		UserID:    user_id_int64,
		Author:    author,
		Tag:       tag,
		Favorited: favoritedParam,
		Limit:     int64(limitInt),
		Offset:    int64(offsetInt),
	})
	if err != nil && err != sql.ErrNoRows {
		h.serverError(w, err)
		return
	}

	res := []map[string]interface{}{}
	for _, v := range articles {
		res = append(res, map[string]interface{}{
			"article": map[string]interface{}{
				"slug":           v.Slug,
				"title":          v.Title,
				"description":    v.Description,
				"tagList":        strings.Split(v.Tags.(string), ","),
				"createdAt":      v.CreatedAt,
				"updatedAt":      v.CreatedAt,
				"favorited":      v.Favorited > 0,
				"favoritesCount": v.FavoritesCount,
				"author": map[string]interface{}{
					"username":  v.User.Username,
					"bio":       v.User.Bio,
					"image":     v.User.Image,
					"following": false,
				},
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

func (h *Handler) FeedArticles(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.serverError(w, err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries

	article, err := qtx.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			h.notFound(w)
			return
		}
		h.serverError(w, err)
		return
	}

	var isFollowing bool
	var isFavorited bool
	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id, ok := claims["user_id"].(float64)
	if ok {
		// check if user_id is following the author
		following, err := h.Queries.IsUserFollowByUserID(r.Context(),
			model.IsUserFollowByUserIDParams{
				UserID:     article.Article.AuthorID,
				FollowerID: int64(user_id),
			})
		if err != nil {
			h.serverError(w, err)
			return
		}
		isFollowing = following > 0
		// check if user_id favorited the article
		favorited, err := h.Queries.IsFavoriteByUserIDAndArticleID(r.Context(),
			model.IsFavoriteByUserIDAndArticleIDParams{
				UserID:    int64(user_id),
				ArticleID: article.Article.ID,
			})
		if err != nil {
			h.serverError(w, err)
			return
		}
		isFavorited = favorited > 0
	}

	err = tx.Commit()
	if err != nil {
		h.serverError(w, err)
		return
	}

	res := map[string]interface{}{
		"article": map[string]interface{}{
			"slug":           article.Article.Slug,
			"title":          article.Article.Title,
			"description":    article.Article.Description,
			"body":           article.Article.Body,
			"tagList":        strings.Split(article.Tag, ","),
			"createdAt":      article.Article.CreatedAt,
			"updatedAt":      article.Article.CreatedAt,
			"favorited":      isFavorited,
			"favoritesCount": 0,
			"author": map[string]interface{}{
				"username":  article.User.Username,
				"bio":       article.User.Bio,
				"image":     article.User.Image,
				"following": isFollowing,
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
