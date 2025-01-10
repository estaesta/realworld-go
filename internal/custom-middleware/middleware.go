package custommiddleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
)

func Verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(ja, TokenFromHeader)(next)
	}
}

func TokenFromHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 6 && strings.ToUpper(bearer[0:5]) == "TOKEN" {
		return bearer[6:]
	}
	return ""
}
