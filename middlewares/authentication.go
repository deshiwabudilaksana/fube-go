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
	Vendor   *Vendor
	jwt.RegisteredClaims
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
	Access   string    `json:"access"`
	Email    string    `json:"email"`
	ExpireAt time.Time `json:"expireAt"`
	Username string    `json:"username"`
	Vendor   *Vendor   `json:"vendor"`
}

type ctxKey string

const userContextKey ctxKey = "user"

func NewAuthContext(claims *JwtClaims) *UserData {
	authCtx := &UserData{
		Access:   string(claims.Access),
		Email:    claims.Email,
		ExpireAt: claims.ExpireAt,
		Username: claims.Username,
		Vendor:   claims.Vendor,
	}

	return authCtx
}

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

		tokenClaims, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JWTSecret, nil
		})

		// if claims, ok := tokenClaims.Claims.(jwt.MapClaims); ok && token.Valid {
		// 	fmt.Println("\n=== Standard Claims (MapClaims) ===")
		// 	for key, value := range claims {
		// 		fmt.Printf("%s: %v\n", key, value)
		// 	}
		// }

		claims := tokenClaims.Claims.(*JwtClaims)

		authContext := NewAuthContext(claims)

		fmt.Println("new auth context >>>", authContext)

		authUser := UserData{
			Access:   authContext.Access,
			Email:    authContext.Email,
			ExpireAt: authContext.ExpireAt,
			Username: authContext.Username,
			Vendor:   authContext.Vendor,
		}

		fmt.Println(authUser)
		ctx := context.WithValue(r.Context(), userContextKey, authUser)
		// ctx := context.WithValue(r.Context(), userContextKey, authContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (UserData, bool) {
	user, ok := ctx.Value(userContextKey).(UserData)
	fmt.Println("user login >>", user)
	return user, ok
}
