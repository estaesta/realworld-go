package main

import (
	"log"
	"net/http"
	"os"

	custommiddleware "github.com/estaesta/realworld-go/internal/custom-middleware"
	"github.com/estaesta/realworld-go/internal/handler"
	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

func (app *application) loadRoutes() {
	errorLog := log.New(
		os.Stderr,
		"ERROR\t",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	queries := model.New(app.DB)
	handler := &handler.Handler{
		ErrorLog: errorLog,
		InfoLog:  infoLog,
		Token:    tokenAuth,
		Queries:  queries,
		DB:       app.DB,
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api", func(r chi.Router) {
		// r.Use(middleware.RequestID)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", handler.Register)
			r.Post("/login", handler.Login)
		})

		r.Group(func(r chi.Router) {
			// Seek, verify and validate JWT tokens
			r.Use(custommiddleware.Verifier(app.token))
			r.Use(jwtauth.Authenticator(app.token))

			// r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			// 	_, claims, _ := jwtauth.FromContext(r.Context())
			// 	w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
			// })
			r.Get("/user", handler.GetUser)
			r.Put("/user", handler.UpdateUser)
			r.Post("/profiles/{username}/follow", handler.Follow)
			r.Delete("/profiles/{username}/follow", handler.Unfollow)
		})

		r.Route("/profiles", func(r chi.Router) {
			r.Use(custommiddleware.Verifier(app.token))
			r.Get("/{username}", handler.GetProfile)
		})
	})

	app.router = r
}
