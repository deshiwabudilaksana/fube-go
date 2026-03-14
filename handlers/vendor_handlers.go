package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/gorilla/mux"
)

type VendorDataInput struct {
	CompanyName string
	Email       string
	Phone       string
	Address     string
	UserId      string
}

/*
Debug purpose handler
*/
func GetUserVendorDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	currentUser, ok := middlewares.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User data not found", http.StatusBadRequest)
		return
	}

	if currentUser.Vendor.ID != vars["id"] {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentUser.Vendor)
}

func AddVendor(w http.ResponseWriter, r *http.Request) {
	// currentUser, ok := middlewares.GetUserFromContext(r.Context())

	// if !ok {
	// 	http.Error(w, "User data not found", http.StatusBadRequest)
	// 	return
	// }

	var input VendorDataInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newVendorRequest := models.Vendor{
		CompanyName: input.CompanyName,
		Email:       input.Email,
		Phone:       input.Phone,
		Address:     input.Address,
		UserID:      &input.UserId,
		IsRemoved:   false,
	}

	newVendor := database.DB.Create(newVendorRequest)

	if newVendor.Error != nil {
		http.Error(w, "Error creating vendor", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newVendorRequest)
}
