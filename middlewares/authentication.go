package middlewares

import (
	"context"
	"fmt"
	"os"

	// "encoding/json"

	"net/http"
	"strings"
	"time"

	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type JwtClaims struct {
	Username string
	Email    string
	Access   models.Role
	ExpireAt time.Time
	Vendor   models.Vendor
}

type Vendor struct {
	Address     string    `json:"address"`
	CompanyName string    `json:"company_name"`
	CreatedAt   time.Time `json:"created_at"`
	Email       string    `json:"email"`
	ID          string    `json:"id"`
	IsRemoved   bool      `json:"is_removed"`
	Phone       string    `json:"phone"`
	UserID      string    `json:"user_id"`
}

type UserData struct {
	Access   string         `json:"access"`
	Email    string         `json:"email"`
	ExpireAt time.Time      `json:"expireAt"`
	Username string         `json:"username"`
	Vendor   map[Vendor]int `json:"vendor"`
}

type ctxKey string

const userContextKey ctxKey = "user"

func Bearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := godotenv.Load(); err != nil {
			http.Error(w, "ENV not found", http.StatusInternalServerError)
		}

		fmt.Println("running auth bearer >>>")
		// Get a single header value
		bearer := r.Header.Get("Authorization")
		splitBearer := strings.Split(bearer, " ")
		JWTSecret := os.Getenv("JWT_SECRET_KEY")

		tokenString := splitBearer[1]
		token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})

		if err != nil {
			http.Error(w, "Unable to verify token", http.StatusFailedDependency)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusFailedDependency)
			return
		}

		decodedToken := token.Claims.(*jwt.MapClaims)

		decodedString := make(map[string]interface{})

		for k, v := range *decodedToken {
			decodedString[k] = v
		}

		expireAt, err := time.Parse(time.RFC3339, decodedString["expireAt"].(string))
		if err != nil {
			http.Error(w, err.Error(), http.StatusFailedDependency)
			return
		}

		fmt.Println("decoded token >>", decodedString)

		var vendorMap map[Vendor]int

		authUser := UserData{
			Access:   decodedString["access"].(string),
			Email:    decodedString["email"].(string),
			ExpireAt: expireAt,
			Username: decodedString["username"].(string),
			Vendor:   vendorMap,
		}
		ctx := context.WithValue(r.Context(), userContextKey, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (UserData, bool) {
	user, ok := ctx.Value(userContextKey).(UserData)
	return user, ok
}
