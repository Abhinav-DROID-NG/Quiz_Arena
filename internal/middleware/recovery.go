package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery is a middleware that recovers from panics, logs the stack trace,
// and returns a 500 Internal Server Error to the client.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
