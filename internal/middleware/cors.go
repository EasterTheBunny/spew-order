package middleware

import (
	"net/http"
)

// SetCORSHeaders is a middleware function that offers CORS headers for an
// OPTION request. This should be used earliest in the base router.
func SetCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		}
	})
}
