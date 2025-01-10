package handler

import (
	"log"

	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/jwtauth/v5"
)

type Handler struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	Token    *jwtauth.JWTAuth
	Queries  *model.Queries
}

// func New(m *model.Model) *Handler {
// 	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
// 	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
// 	h := &Handler{}
// 	h.ErrorLog = errorLog
// 	h.InfoLog = infoLog
// 	return h
// }
