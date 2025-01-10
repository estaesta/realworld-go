package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	type request struct {
		User struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}
	var req = request{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.InfoLog.Println(err)
		h.clientError(w, http.StatusBadRequest)
		return
	}

	// validate
	if req.User.Username == "" || req.User.Email == "" ||
		len(req.User.Password) < 8 {
		h.InfoLog.Println(req.User)
		h.clientError(w, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.User.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		h.serverError(w, err)
		return
	}

	q := h.Queries

	u, err := q.CreateUser(r.Context(), model.CreateUserParams{
		Email:    req.User.Email,
		Password: string(hashedPassword),
		Username: req.User.Username,
	})

	// check if error unique constraint
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		if sqliteErr.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) {
			h.clientError(w, http.StatusBadRequest)
			return
		}
		h.InfoLog.Println("Database error: ", err)
		h.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	res := map[string]interface{}{
		"user": map[string]interface{}{
			"email":    u.Email,
			"token":    nil,
			"username": u.Username,
			"bio":      u.Bio,
			"image":    u.Image,
		},
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		h.serverError(w, err)
		return
	}

	w.Write(resJson)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	type request struct {
		User struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}
	var req = request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	q := h.Queries
	u, err := q.GetUserByEmail(r.Context(), req.User.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			h.clientError(w, http.StatusUnauthorized)
			return
		}
		h.InfoLog.Println("Database error: ", err)
		h.serverError(w, err)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(req.User.Password),
	)
	if err != nil {
		h.clientError(w, http.StatusUnauthorized)
		return
	}

	_, tokenString, _ := h.Token.Encode(map[string]interface{}{"user_id": u.ID})
	res := map[string]interface{}{
		"user": map[string]interface{}{
			"email":    u.Email,
			"token":    tokenString,
			"username": u.Username,
			"bio":      u.Bio,
			"image":    u.Image,
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
