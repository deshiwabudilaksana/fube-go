package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetCustomerHandlers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.Println("vars >>>", vars["id"])

	currentUser, ok := middlewares.GetUserFromContext(r.Context())

	if !ok {
		http.Error(w, "User data not found", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	// database.DB.Where("vendor_id = ?", currentUser.Vendor.ID).Find(&allCustomer)

	id, err := uuid.Parse(vars["id"])

	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	database.DB.Where(&models.Customer{ID: id, VendorID: currentUser.Vendor.ID}).First(&customer)

	log.Println("customer found >>", customer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func GetAllCustomersHandlers(w http.ResponseWriter, r *http.Request) {
	customers, err := models.GetAllCustomers()
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "No customers found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving customers", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

type CustomerInput struct {
	Name      string
	Birthdate time.Time
	Phone     string
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middlewares.GetUserFromContext(r.Context())

	if !ok {
		http.Error(w, "User data not found", http.StatusBadRequest)
		return
	}
	// This is how to get data from multipart form request
	// if err := r.ParseForm(); err != nil {
	// 	http.Error(w, "Can not send empty request", http.StatusBadRequest)
	// }

	// customer := CustomerInput{
	// 	Name:      r.Form.Get("name"),
	// 	Birthdate: r.Form.Get("birthdate"),
	// 	Phone:     r.Form.Get("phone"),
	// }

	// log.Println("Customer data parsed >>>", customer)

	//To do: add authentication

	var input CustomerInput

	// Decode JSON body into struct
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Println("Customer data decoder >>>", input)

	// newCustomer := database.DB.Create(customer)
	newCustomer := models.Customer{
		FullName:  input.Name,
		Birthdate: input.Birthdate,
		Phone:     input.Phone,
		Points:    0,
		VendorID:  currentUser.Vendor.ID,
		IsRemoved: false,
	}

	result := database.DB.Create(&newCustomer)

	log.Println("Result >>", result)

	if result.Error != nil {
		http.Error(w, "Error creating customer", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newCustomer)
}

func TestGetRequestBody(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Can not send empty request", http.StatusBadRequest)
	}

	customer := CustomerInput{
		Name: r.Form.Get("name"),
		// Birthdate: r.Form.Get("birthdate"),
		Phone: r.Form.Get("phone"),
	}

	log.Println("get request >> ", customer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
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

type StringKey struct{}

// type requestDataKey struct{}

func GetCustomer(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middlewares.GetUserFromContext(r.Context())

	if !ok {
		http.Error(w, "User data not found", http.StatusBadRequest)
		return
	}

	var allCustomer models.Customer
	database.DB.Where("vendor_id = ?", currentUser.Vendor.ID).Find(&allCustomer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allCustomer)
}
