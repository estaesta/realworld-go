package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/estaesta/realworld-go/internal/util"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	router http.Handler
	port   int
	DB     *sql.DB
	token  *jwtauth.JWTAuth
}

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil) // replace with secret key

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

func main() {
	// Connect to the database
	db, err := sql.Open("sqlite3", "./db/realworld.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//ping db
	if err := db.Ping(); err != nil {
		panic(err)
	}

	app := &application{
		port:  8080,
		DB:    db,
		token: tokenAuth,
	}
	fmt.Println("loading routes")
	app.loadRoutes()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.port),
		Handler: app.router,
	}

	done := make(chan bool)
	go util.GracefulShutdown(server, done)

	fmt.Printf("Server started on port %d\n", app.port)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	<-done
	fmt.Println("Server stopped")
}
