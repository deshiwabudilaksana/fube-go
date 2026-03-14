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

	/*
		Super User only, can create new owner
	*/
	apiRoute.Handle("/create_owner", middlewares.SuperUserValidation(middlewares.UserTypeOwnerContext(http.HandlerFunc(handlers.CreateUser))))
	apiRoute.Handle("/create_vendor", middlewares.SuperUserValidation(http.HandlerFunc(handlers.AddVendor)))

	// Yield Recipe Planning Handlers
	apiRoute.HandleFunc("/menus/{id}/cost", handlers.GetMenuCostHandler).Methods("GET")
	apiRoute.HandleFunc("/planning/production", handlers.PlanProductionHandler).Methods("POST")
	apiRoute.HandleFunc("/reports/external", handlers.GetExternalReportingHandler).Methods("GET")
	apiRoute.HandleFunc("/export/costs", handlers.DownloadCostReportHandler).Methods("GET")
	apiRoute.HandleFunc("/import/csv", handlers.ImportCSVHandler).Methods("POST")

	// Auth Handlers (Public)
	r.HandleFunc("/login", handlers.ShowLogin).Methods("GET")
	r.HandleFunc("/login", handlers.PostLogin).Methods("POST")
	r.HandleFunc("/register", handlers.ShowRegister).Methods("GET")
	r.HandleFunc("/register", handlers.PostRegister).Methods("POST")
	r.HandleFunc("/logout", handlers.Logout).Methods("POST")

	// Landing Page (Public)
	r.HandleFunc("/", handlers.RenderLandingPage).Methods("GET")

	// Frontend Handlers (HTMX - Protected)
	fe := r.PathPrefix("/").Subrouter()
	fe.Use(middlewares.Bearer)

	fe.HandleFunc("/yield-planning", handlers.RenderYieldPlanning).Methods("GET")
	fe.HandleFunc("/yield-planning/update", handlers.UpdateYieldPlanning).Methods("POST")
	fe.HandleFunc("/import", handlers.RenderImport).Methods("GET")
	fe.HandleFunc("/inventory", handlers.GetInventoryDashboard).Methods("GET")
	fe.HandleFunc("/inventory/stock-take", handlers.PostStockTake).Methods("POST")

	// Mapping UI Handlers
	fe.HandleFunc("/mapping", handlers.GetMappingDashboard).Methods("GET")
	fe.HandleFunc("/mapping/search-yields", handlers.SearchYields).Methods("GET")
	fe.HandleFunc("/mapping/update", handlers.UpdateMenuMapping).Methods("POST")
	fe.HandleFunc("/mapping/picker", handlers.RenderRecipePicker).Methods("GET")
	fe.HandleFunc("/export/costs", handlers.DownloadCostReportHandler).Methods("GET")

	// KDS Handlers
	fe.HandleFunc("/kds", handlers.GetKDSBoard).Methods("GET")
	fe.HandleFunc("/kds/updates", handlers.GetKDSBoard).Methods("GET")
	fe.HandleFunc("/kds/orders/{id}/status", handlers.UpdateOrderStatus).Methods("POST")
	fe.HandleFunc("/kds/new-count", handlers.GetNewOrdersCount).Methods("GET")

	// Supplier Handlers
	fe.HandleFunc("/suppliers", handlers.GetSupplierDashboard).Methods("GET")
	fe.HandleFunc("/suppliers/generate-po", handlers.PostGeneratePO).Methods("POST")

	return r
}
