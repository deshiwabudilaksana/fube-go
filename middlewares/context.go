package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/deshiwabudilaksana/fube-go/models"
)

type StringKey struct{}

func TestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), StringKey{}, "userId")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type CustomerInput struct {
	Name      string
	Birthdate time.Time
	Phone     string
	Role      models.Role
}

type UserDataInput struct {
	Username string
	Email    string
	Phone    string
	VendorId string
	Role     models.Role
}

func UserTypeOwnerContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input UserDataInput

		// Decode JSON body into struct
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		input.Role = "OWNER"

		ctx := context.WithValue(r.Context(), StringKey{}, input)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}
