package router

import (
	"github.com/deshiwabudilaksana/fube-go/handlers"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	preAuth := r.PathPrefix("/preauth").Subrouter()
	preAuth.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	preAuth.HandleFunc("/signup", handlers.CreateUser).Methods("POST")

	apiRoute := r.PathPrefix("/api").Subrouter()
	apiRoute.Use(middlewares.Bearer)

	// User Handlers
	// To do: fix input params
	apiRoute.HandleFunc("/users/{id}", handlers.GetUserHandlers).Methods("GET")
	apiRoute.HandleFunc("/users", handlers.GetAllUsersHandlers).Methods("GET")

	// Customer Handlers
	apiRoute.HandleFunc("/customers", handlers.GetAllCustomersHandlers).Methods("GET")
	// To do: fix input params
	apiRoute.HandleFunc("/customers/{id}", handlers.GetCustomerHandlers).Methods("GET")
	apiRoute.HandleFunc("/customers", handlers.AddCustomer).Methods("POST")

	apiRoute.HandleFunc("/test", handlers.TestGetRequestBody).Methods("POST")
	apiRoute.HandleFunc("/test_bearer", handlers.GetCustomer).Methods("GET")

	return r
}
