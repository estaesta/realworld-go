package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	custommiddleware "github.com/estaesta/realworld-go/internal/custom-middleware"
	"github.com/estaesta/realworld-go/internal/handler"
	"github.com/estaesta/realworld-go/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://realworld.ddpaud.my.id*", "https://estaesta.github.io*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		res := map[string]interface{}{
			"status": "UP",
		}
		resJson, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resJson)
	})

	r.Route("/api", func(r chi.Router) {

		r.Route("/users", func(r chi.Router) {
			r.Post("/", handler.Register)
			r.Post("/login", handler.Login)
		})

		r.Get("/tags", handler.GetTags)

		// Auth required
		r.Group(func(r chi.Router) {
			// Seek, verify and validate JWT tokens
			r.Use(custommiddleware.Verifier(app.token))
			r.Use(jwtauth.Authenticator(app.token))

			r.Get("/user", handler.GetUser)
			r.Put("/user", handler.UpdateUser)

			r.Post("/profiles/{username}/follow", handler.Follow)
			r.Delete("/profiles/{username}/follow", handler.Unfollow)

			r.Get("/articles/feed", handler.FeedArticles)
			r.Post("/articles", handler.CreateArticle)
			r.Put("/articles/{slug}", handler.UpdateArticle)
			r.Delete("/articles/{slug}", handler.DeleteArticle)

			r.Post("/articles/{slug}/comments", handler.AddComment)
			r.Delete("/articles/{slug}/comments/{id}", handler.DeleteComment)

			r.Post("/articles/{slug}/favorite", handler.FavoriteArticle)
			r.Delete("/articles/{slug}/favorite", handler.UnfavoriteArticle)
		})

		// Auth optional
		r.Group(func(r chi.Router) {
			r.Use(custommiddleware.Verifier(app.token))

			r.Get("/profiles/{username}", handler.GetProfile)

			r.Get("/articles", handler.ListArticles)
			r.Get("/articles/{slug}", handler.GetArticle)

			r.Get("/articles/{slug}/comments", handler.GetComments)
		})

	})

	app.router = r
}
