package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type ReturnObject struct {
	Status bool
	Token  string
}

func GetUserHandlers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	user, err := models.GetUser(vars["id"]) // Convert back to uint for GORM
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GetAllUsersHandlers(w http.ResponseWriter, r *http.Request) {
	result, err := models.GetAllUsers()
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "No users found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type userLogin struct {
	UserName   string
	Password   string
	VendorName string
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	PepperString := os.Getenv("HASH_PEPPER")

	log.Println("Authenticating user...")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "False request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var requestBody userLogin
	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userRequest := userLogin{
		UserName:   requestBody.UserName,
		Password:   requestBody.Password,
		VendorName: requestBody.VendorName,
	}

	log.Println("Login attempt for user:", userRequest.UserName)

	var vendor models.Vendor
	validVendor := database.DB.Where("company_name = ?", userRequest.VendorName).First(&vendor)

	if validVendor.Error != nil {
		http.Error(w, "Invalid vendor", http.StatusBadRequest)
		return
	}

	var foundUser models.User
	result := database.DB.Where("username = ?", userRequest.UserName).Where("vendor_id = ?", vendor.ID).Find(&foundUser)

	if result.Error != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userRequest.Password+PepperString))
	if err != nil {
		http.Error(w, "Wrong password", http.StatusBadRequest)
		return
	}

	claim := jwt.MapClaims{
		"username": foundUser.Username,
		"email":    foundUser.Email,
		"access":   foundUser.Role,
		"expireAt": time.Now().Add(24 * time.Hour).Unix(), // Changed to Unix for standard JWT
		"vendor":   vendor,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	signedToken, err := token.SignedString([]byte(JWTSecret))

	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	returnObject := ReturnObject{
		Status: true,
		Token:  signedToken,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(returnObject); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type newUser struct {
	UserName   string
	Password   string
	Email      string
	Phone      string
	Role       models.Role
	VendorName string
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "False request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var requestBody newUser
	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newUserRequest := newUser{
		UserName:   requestBody.UserName,
		Password:   requestBody.Password,
		Email:      requestBody.Email,
		Phone:      requestBody.Phone,
		Role:       requestBody.Role,
		VendorName: requestBody.VendorName,
	}

	var foundVendor models.Vendor
	result := database.DB.Where("company_name = ?", newUserRequest.VendorName).First(&foundVendor)

	if result.Error != nil {
		http.Error(w, "Invalid vendor", http.StatusBadRequest)
		return
	}

	// NOTE: In a real system, you'd hash the password here before saving.
	// This code seems to be creating a user directly from the request.
	if err := database.DB.Create(&newUserRequest).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	claim := jwt.MapClaims{
		"username": newUserRequest.UserName,
		"email":    newUserRequest.Email,
		"access":   newUserRequest.Role,
		"expireAt": time.Now().Add(24 * time.Hour).Unix(),
		"vendor":   foundVendor.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim) // Fixed from ES256 to match LoginUser
	signedToken, err := token.SignedString([]byte(JWTSecret))

	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	returnObject := ReturnObject{
		Status: true,
		Token:  signedToken,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(returnObject); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
