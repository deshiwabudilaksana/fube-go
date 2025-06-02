package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/gorilla/mux"
)

func GetCustomerHandlers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	customer, err := models.GetCustomer(vars["id"])
	if err != nil {
		if err.Error() == "customer not found" {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving customer data", http.StatusInternalServerError)
		}
		return
	}

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
	Birthdate string
	Phone     string
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Can not send empty request", http.StatusBadRequest)
	}

	customer := CustomerInput{
		Name:      r.Form.Get("name"),
		Birthdate: r.Form.Get("birthdate"),
		Phone:     r.Form.Get("phone"),
	}

	//To do: add authentication

	newCustomer := database.DB.Create(customer)

	if newCustomer.Error != nil {
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
		Name:      r.Form.Get("name"),
		Birthdate: r.Form.Get("birthdate"),
		Phone:     r.Form.Get("phone"),
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
	// fmt.Println("user access:", currentUser.Access)
	// fmt.Println("user email:", currentUser.Email)
	// fmt.Println("user ExpireAt:", currentUser.ExpireAt)
	fmt.Println("user Vendor:", currentUser.Vendor)

	// const userAccess := currentUser.Access
	if !ok {
		http.Error(w, "User data not found", http.StatusBadRequest)
		return
	}

	var allCustomer models.Customer
	database.DB.Where("vendor_id = ?", currentUser.Vendor).Find(&allCustomer)

	fmt.Println("all cust >>", allCustomer)
}

// func TestContext(w http.ResponseWriter, r *http.Request) {
// 	stringContext, ok := r.Context().Value(requestDataKey{}).(UserData)
// 	if !ok {
// 		http.Error(w, "User data not found", http.StatusInternalServerError)
// 		return
// 	}
// 	fmt.Println("Str ctx:", stringContext) // Output: 123
// }
