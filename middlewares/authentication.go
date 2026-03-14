package middlewares

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	UserID   string `json:"user_id"`
	VendorID string `json:"vendor_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Vendor struct {
	ID          string    `json:"id"`
	Address     string    `json:"address,omitempty"`
	CompanyName string    `json:"company_name,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	Email       string    `json:"email,omitempty"`
	IsRemoved   bool      `json:"is_removed,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
}

type UserData struct {
	Access   string    `json:"access"`
	Email    string    `json:"email"`
	ExpireAt time.Time `json:"expireAt"`
	Username string    `json:"username"`
	Vendor   *Vendor   `json:"vendor"`
	UserID   string    `json:"user_id"`
}

type ctxKey string

const userContextKey ctxKey = "user"

func Bearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("running auth bearer >>>")

		var tokenString string

		// 1. Check Authorization header
		bearer := r.Header.Get("Authorization")
		if bearer != "" {
			splitBearer := strings.Split(bearer, " ")
			if len(splitBearer) >= 2 {
				tokenString = splitBearer[1]
			}
		}

		// 2. Check jwt cookie if header is missing or empty
		if tokenString == "" {
			cookie, err := r.Cookie("jwt")
			if err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			// If it's a browser request (HTMX or regular GET), redirect to login
			if r.Header.Get("HX-Request") == "true" || strings.Contains(r.Header.Get("Accept"), "text/html") {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		JWTSecret := os.Getenv("JWT_SECRET_KEY")
		if JWTSecret == "" {
			log.Println("JWT_SECRET_KEY is not set")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(JWTSecret), nil
		})

		if err != nil || !token.Valid {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		claims, ok := token.Claims.(*JwtClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Check expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		authUser := UserData{
			Access:   claims.Role,
			Email:    claims.Email,
			Username: claims.Username,
			UserID:   claims.UserID,
			Vendor: &Vendor{
				ID: claims.VendorID,
			},
		}

		if claims.ExpiresAt != nil {
			authUser.ExpireAt = claims.ExpiresAt.Time
		}

		ctx := context.WithValue(r.Context(), userContextKey, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (UserData, bool) {
	user, ok := ctx.Value(userContextKey).(UserData)
	return user, ok
}
