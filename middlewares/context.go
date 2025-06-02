package middlewares

import (
	"context"
	"net/http"
)

type StringKey struct{}

func TestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), StringKey{}, "userId")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
