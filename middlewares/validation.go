package middlewares

import (
	"fmt"
	"log"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/models"
)

func ValidateUserVendorFromContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("validate user vendor >>>")
		currentUser, ok := GetUserFromContext(r.Context())

		if !ok {
			http.Error(w, "User data not found", http.StatusBadRequest)
			return
		}

		var validVendor models.Vendor
		database.DB.Where("ID = ?", currentUser.Vendor.ID).Find(&validVendor)

		fmt.Println("valid vendor >>")
		fmt.Println(validVendor)

		next.ServeHTTP(w, r)
	})
}

func SpecificRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[SpecificRouteMiddleware] Executing specific logic for this route.")
		next.ServeHTTP(w, r)
	})
}
