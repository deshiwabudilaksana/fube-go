package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
	if err := godotenv.Load(); err != nil {
		http.Error(w, "ENV not found", http.StatusInternalServerError)
	}

	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	PepperString := os.Getenv("HASH_PEPPER")

	fmt.Println("jwt secret >>", JWTSecret)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "False request", http.StatusBadRequest)
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	var requestBody userLogin
	json.Unmarshal(body, &requestBody)

	userRequest := userLogin{
		UserName:   requestBody.UserName,
		Password:   requestBody.Password,
		VendorName: requestBody.VendorName,
	}

	log.Println(userRequest)

	var vendor models.Vendor
	validVendor := database.DB.Where("company_name = ?", userRequest.VendorName).First(&vendor)

	log.Println("found vendor", vendor.ID)

	if validVendor.Error != nil {
		http.Error(w, "Invalid vendor", http.StatusBadRequest)
		return
	}

	// to do: parse gorm result
	var foundUser models.User
	result := database.DB.Where("username = ?", userRequest.UserName).Where("vendor_id = ?", vendor.ID).Find(&foundUser)

	log.Println("found user", foundUser)

	if result.Error != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	log.Println("user password >> ", foundUser.Password)
	log.Println("plain text >> ", userRequest.Password+PepperString)
	log.Println("pepper string >> ", PepperString)

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userRequest.Password+PepperString))

	if err != nil {
		http.Error(w, "Wrong password", http.StatusBadRequest)
		return
	}

	claim := jwt.MapClaims{
		"username": foundUser.Username,
		"email":    foundUser.Email,
		"access":   foundUser.Role,
		"expireAt": time.Now().Add(24 * time.Hour),
		"vendor":   vendor,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	// log.Println("token >> ", token.Claims)
	log.Print("string secret >> ", []byte("JWTSecret"))

	signedToken, err := token.SignedString([]byte(JWTSecret))

	// log.Println("JWT Secret >> ", signedToken)
	// log.Println("error signed token >> ", err)

	if err != nil {
		http.Error(w, "Wrong jwt secret", http.StatusFailedDependency)
		return
	}

	returnObject := ReturnObject{
		Status: true,
		Token:  signedToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(returnObject)
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
	if err := godotenv.Load(); err != nil {
		http.Error(w, "ENV not found", http.StatusInternalServerError)
	}

	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "False request", http.StatusBadRequest)
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	var requestBody newUser
	json.Unmarshal(body, &requestBody)

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

	newUser := database.DB.Create(newUserRequest)

	if newUser.Error != nil {
		http.Error(w, "Error creating user", http.StatusConflict)
		return
	}

	claim := jwt.MapClaims{
		"username": newUserRequest.UserName,
		"email":    newUserRequest.Email,
		"access":   newUserRequest.Role,
		"expireAt": time.Now().Add(24 * time.Hour),
		"vendor":   foundVendor.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claim)

	signedToken, err := token.SignedString([]byte(JWTSecret))

	if err != nil {
		http.Error(w, "Wrong jwt secret", http.StatusFailedDependency)
		return
	}

	returnObject := ReturnObject{
		Status: true,
		Token:  signedToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(returnObject)
}
