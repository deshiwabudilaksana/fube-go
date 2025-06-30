package router

import (
	"net/http"

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
	apiRoute.Handle("/users/{id}", middlewares.ValidateUserVendorFromContext(http.HandlerFunc(handlers.GetUserHandlers))).Methods("GET")
	apiRoute.HandleFunc("/users", handlers.GetAllUsersHandlers).Methods("GET")

	// Customer Handlers
	apiRoute.HandleFunc("/customers", handlers.GetAllCustomersHandlers).Methods("GET")
	apiRoute.Handle("/customers/{id}", middlewares.ValidateUserVendorFromContext(http.HandlerFunc(handlers.GetCustomerHandlers))).Methods("GET")
	// To do: fix input params
	apiRoute.Handle("/customers", middlewares.ValidateUserVendorFromContext(http.HandlerFunc(handlers.AddCustomer))).Methods("POST")

	apiRoute.HandleFunc("/test", handlers.TestGetRequestBody).Methods("POST")
	apiRoute.HandleFunc("/test_bearer", handlers.GetCustomer).Methods("GET")
	apiRoute.HandleFunc("/test/middleware", handlers.GetCustomer).Methods("GET")

	//Validate Vendor User Handlers
	validVendorRoute := *apiRoute
	validVendorRoute.Use(middlewares.ValidateUserVendorFromContext)

	/**
	This router use all implemented middleware from apiRoute sub route and implement the specific handler
	*/
	validVendorRoute.HandleFunc("/validate_vendor", handlers.GetCustomer).Methods("GET")

	/**
	This router specify which middleware to be implemented on this single route
	*/
	r.Handle("/check_vendor", middlewares.Bearer(middlewares.ValidateUserVendorFromContext(http.HandlerFunc(handlers.GetCustomer)))).Methods("GET")

	/**
	Both of those 2 route will return the same result and use same middlewares
	*/

	apiRoute.Handle("/combined_vendor", middlewares.ValidateUserVendorFromContext(http.HandlerFunc(handlers.GetCustomer))).Methods("GET")
	/*
		Combined those 2 method, using subrouter for authentication and implement middlewares for specific routes
	*/

	return r
}
