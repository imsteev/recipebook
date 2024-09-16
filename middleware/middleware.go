package middleware

import (
	"net/http"
)

func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "HX-Request")
		w.Header().Add("Cache-Control", "no-cache, no-store")
		next.ServeHTTP(w, r)
	})
}
