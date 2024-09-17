package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type LoggedInUserCtxKey struct{}

// problem i'm trying to solve: make sure that the back button doesn't reveal
// private data. the example is: logged in user can see private recipes, but upon
// logout, they can't go back to the recipes and see them. currently that's possible.
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "HX-Request")
		w.Header().Add("Cache-Control", "no-cache, no-store")
		next.ServeHTTP(w, r)
	})
}

func RequireAuth(store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sesh, err := store.Get(r, "sesh")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if sesh.Values["loggedInUserID"] == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, LoggedInUserCtxKey{}, sesh.Values["loggedInUserID"])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
