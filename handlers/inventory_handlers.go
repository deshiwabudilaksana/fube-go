package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
)

// GetInventoryDashboard renders the HTMX inventory view.
// Tiger Style: Explicit context management and robust error logging.
func GetInventoryDashboard(w http.ResponseWriter, r *http.Request) {
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

	// Default range for theoretical usage (last 30 days)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	auditData, err := models.GetInventoryAudit(r.Context(), database, vendorID, storeID, startTime, endTime)
	if err != nil {
		log.Printf("Error fetching inventory audit for vendor %s: %v", vendorID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/inventory_fragment.html",
	)
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		AuditItems []models.InventoryAuditResult
		StartTime  time.Time
		EndTime    time.Time
	}{
		AuditItems: auditData,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// PostStockTake processes physical count updates via HTMX.
// Tiger Style: Robust input validation and immediate feedback loop.
func PostStockTake(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	materialID := r.FormValue("material_id")
	quantity, err := strconv.ParseFloat(r.FormValue("quantity"), 64)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	user, ok := middlewares.GetUserFromContext(r.Context())
	vendorID := "demo-vendor"
	storeID := "demo-store"
	userName := "system"
	if ok {
		if user.Vendor != nil {
			vendorID = user.Vendor.ID
		}
		userName = user.Username
	}

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	err = models.UpdatePhysicalStock(r.Context(), database, vendorID, storeID, materialID, quantity, userName)
	if err != nil {
		log.Printf("Error updating physical stock for material %s: %v", materialID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Re-render the fragment to show updated values and calculated variance
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)
	auditData, err := models.GetInventoryAudit(r.Context(), database, vendorID, storeID, startTime, endTime)
	if err != nil {
		log.Printf("Error fetching refreshed audit: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		AuditItems []models.InventoryAuditResult
		StartTime  time.Time
		EndTime    time.Time
	}{
		AuditItems: auditData,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	tmpl, err := template.ParseFiles("templates/inventory_fragment.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "main", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
