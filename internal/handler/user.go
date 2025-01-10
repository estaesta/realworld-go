package handler

import (
	"encoding/json"
	"net/http"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	user_id := int64(claims["user_id"].(float64))

	u, err := h.Queries.GetUserByID(r.Context(), user_id)
	if err != nil {
		h.ErrorLog.Println(err)
		return
	}

	tokenString := r.Header.Get("Authorization")[7:]

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

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		User struct {
			Email    *string `json:"email"`
			Username *string `json:"username"`
			Password *string `json:"password"`
			Image    *string `json:"image"`
			Bio      *string `json:"bio"`
		} `json:"user"`
	}
	var req = request{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.clientError(w, http.StatusBadRequest)
		return
	}

	_, claim, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.clientError(w, http.StatusUnauthorized)
		return
	}

	id := int64(claim["user_id"].(float64))
	token := r.Header.Get("Authorization")[7:]

	param := model.UpdateUserByIDParams{ID: id}
	if req.User.Email != nil {
		if *req.User.Email == "" {
			h.clientError(w, http.StatusBadRequest)
			return
		}
		param.Email = req.User.Email
	}
	if req.User.Username != nil {
		if *req.User.Username == "" {
			h.clientError(w, http.StatusBadRequest)
			return
		}
		param.Username = req.User.Username
	}
	if req.User.Password != nil {
		if len(*req.User.Password) < 8 {
			h.clientError(w, http.StatusBadRequest)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(*req.User.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			h.serverError(w, err)
			return
		}
		hashedPasswordString := string(hashedPassword)
		param.Password = &hashedPasswordString
	}
	if req.User.Image != nil {
		param.Image = req.User.Image
	}
	if req.User.Bio != nil {
		param.Bio = req.User.Bio
	}

	user, err := h.Queries.UpdateUserByID(r.Context(), param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"user": map[string]interface{}{
			"email":    user.Email,
			"token":    token,
			"username": user.Username,
			"bio":      user.Bio,
			"image":    user.Image,
		},
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}
