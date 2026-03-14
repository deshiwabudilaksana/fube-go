package middlewares

import (
	"log"
	"net/http"
	"os"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/models"
)

func ValidateUserVendorFromContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("validate user vendor >>>")
		currentUser, ok := GetUserFromContext(r.Context())

		if !ok {
			http.Error(w, "User data not found", http.StatusBadRequest)
			return
		}

		var validVendor models.Vendor
		database.DB.Where("ID = ?", currentUser.Vendor.ID).Find(&validVendor)

		log.Printf("valid vendor >> %+v", validVendor)

		next.ServeHTTP(w, r)
	})
}

func SpecificRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[SpecificRouteMiddleware] Executing specific logic for this route.")
		next.ServeHTTP(w, r)
	})
}

func SuperUserValidation(next http.Handler) http.Handler {
	// Load the super vendor ID once when the function is first called
	superVendorID := os.Getenv("SUPER_VENDOR_ID")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := GetUserFromContext(r.Context())

		if !ok {
			http.Error(w, "User data not found in context", http.StatusUnauthorized)
			return
		}

		if currentUser.Vendor.ID != superVendorID {
			http.Error(w, "Forbidden: Insufficient privileges", http.StatusForbidden)
			return
		}

		// User is authenticated as a super user, proceed to next handler
		next.ServeHTTP(w, r)
	})
}
