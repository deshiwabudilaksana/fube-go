package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
)

// GetSupplierDashboard renders the HTMX supplier management view.
// Tiger Style: Explicit multi-tenancy enforcement and robust error handling.
func GetSupplierDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.GetUserFromContext(r.Context())
	vendorID := "demo-vendor"
	storeID := "demo-store"
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		log.Printf("Error getting MongoDB database: %v", err)
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	suppliers, err := models.GetSuppliers(r.Context(), database, vendorID)
	if err != nil {
		log.Printf("Error fetching suppliers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pendingOrders, err := models.GetPurchaseOrders(r.Context(), database, vendorID, storeID)
	if err != nil {
		log.Printf("Error fetching purchase orders: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	suggestedOrders, err := models.CalculateSuggestedOrders(r.Context(), database, vendorID, storeID)
	if err != nil {
		log.Printf("Error calculating suggestions: %v", err)
		// We don't fail the whole dashboard if suggestions fail, but we log it.
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/supplier_fragment.html",
	)
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Suppliers       []models.SupplierDoc
		PendingOrders   []models.PurchaseOrderDoc
		SuggestedOrders []models.SuggestedOrderItem
	}{
		Suppliers:       suppliers,
		PendingOrders:   pendingOrders,
		SuggestedOrders: suggestedOrders,
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// PostGeneratePO creates draft POs from current suggestions.
// Tiger Style: Idempotent-like behavior and clear success feedback.
func PostGeneratePO(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.GetUserFromContext(r.Context())
	vendorID := "demo-vendor"
	storeID := "demo-store"
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	suggestions, err := models.CalculateSuggestedOrders(r.Context(), database, vendorID, storeID)
	if err != nil {
		log.Printf("Error calculating suggestions for PO generation: %v", err)
		http.Error(w, "Failed to calculate suggestions", http.StatusInternalServerError)
		return
	}

	if len(suggestions) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<div class='p-4 mb-4 text-sm text-yellow-800 rounded-lg bg-yellow-50'>No suggestions available to generate POs.</div>"))
		return
	}

	err = models.GeneratePOFromSuggestions(r.Context(), database, vendorID, storeID, suggestions)
	if err != nil {
		log.Printf("Error generating POs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<div class='p-4 mb-4 text-sm text-green-800 rounded-lg bg-green-50'>Draft Purchase Orders generated successfully!</div>"))
}
