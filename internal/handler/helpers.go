package handler

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/mattn/go-sqlite3"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (h *Handler) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	h.ErrorLog.Println(trace)

	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (h *Handler) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (h *Handler) notFound(w http.ResponseWriter) {
	h.clientError(w, http.StatusNotFound)
}

// ConstraintCheck is a helper function to check if the error is a constraint error
func (h *Handler) ConstraintCheck(err error, w http.ResponseWriter) {
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
}
