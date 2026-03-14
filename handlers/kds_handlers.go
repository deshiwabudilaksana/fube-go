package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/gorilla/mux"
)

// GetKDSBoard renders the Kitchen Display System board.
// Tiger Style: Robust error handling and context-aware database access.
func GetKDSBoard(w http.ResponseWriter, r *http.Request) {
	vendorID := getVendorID(r)
	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		log.Printf("KDS: Database connection error: %v", err)
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	orders, err := models.GetKDSOrders(r.Context(), database, vendorID)
	if err != nil {
		log.Printf("KDS: Error fetching orders for vendor %s: %v", vendorID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"timeAgo": func(t time.Time) string {
			duration := time.Since(t)
			if duration.Minutes() < 1 {
				return "Just now"
			}
			return fmt.Sprintf("%d mins ago", int(duration.Minutes()))
		},
		"isOld": func(t time.Time) bool {
			return time.Since(t).Minutes() > 15
		},
		"getNextStatus": models.GetNextStatus,
		"getStatusLabel": func(s models.OrderStatus) string {
			switch s {
			case models.StatusPending:
				return "Confirm Order"
			case models.StatusConfirmed:
				return "Start Preparing"
			case models.StatusPreparing:
				return "Mark Ready"
			default:
				return "Complete"
			}
		},
	}

	tmpl, err := template.New("kds_fragment.html").Funcs(funcMap).ParseFiles(
		"templates/layout.html",
		"templates/kds_fragment.html",
	)
	if err != nil {
		log.Printf("KDS: Template parse error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Orders []models.OrderDoc
	}{
		Orders: orders,
	}

	// If it's an HTMX request for updates, we might only want the fragment
	if r.Header.Get("HX-Request") == "true" && r.URL.Path == "/kds/updates" {
		err = tmpl.ExecuteTemplate(w, "kds_board", data)
	} else {
		err = tmpl.ExecuteTemplate(w, "layout.html", data)
	}

	if err != nil {
		log.Printf("KDS: Template execution error: %v", err)
	}
}

// UpdateOrderStatus handles status transitions from the KDS board.
// Tiger Style: Atomic status transitions and immediate UI feedback via HTMX.
func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]
	nextStatus := models.OrderStatus(r.FormValue("status"))

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	err = models.UpdateOrderStatus(r.Context(), database, orderID, nextStatus)
	if err != nil {
		log.Printf("KDS: Error updating order %s to %s: %v", orderID, nextStatus, err)
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	// Re-fetch orders and re-render the board for HTMX swap
	vendorID := getVendorID(r)
	orders, err := models.GetKDSOrders(r.Context(), database, vendorID)
	if err != nil {
		log.Printf("KDS: Error fetching orders after update: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"timeAgo": func(t time.Time) string {
			duration := time.Since(t)
			if duration.Minutes() < 1 {
				return "Just now"
			}
			return fmt.Sprintf("%d mins ago", int(duration.Minutes()))
		},
		"isOld": func(t time.Time) bool {
			return time.Since(t).Minutes() > 15
		},
		"getNextStatus": models.GetNextStatus,
		"getStatusLabel": func(s models.OrderStatus) string {
			switch s {
			case models.StatusPending:
				return "Confirm Order"
			case models.StatusConfirmed:
				return "Start Preparing"
			case models.StatusPreparing:
				return "Mark Ready"
			default:
				return "Complete"
			}
		},
	}

	tmpl, err := template.New("kds_fragment.html").Funcs(funcMap).ParseFiles(
		"templates/kds_fragment.html",
	)
	if err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Orders []models.OrderDoc
	}{
		Orders: orders,
	}

	err = tmpl.ExecuteTemplate(w, "kds_board", data)
	if err != nil {
		log.Printf("KDS: Template execution error: %v", err)
	}
}

// GetNewOrdersCount returns the count of PENDING orders for HTMX polling.
// Tiger Style: Lightweight response for high-frequency polling.
func GetNewOrdersCount(w http.ResponseWriter, r *http.Request) {
	vendorID := getVendorID(r)
	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	count, err := models.GetNewOrdersCount(r.Context(), database, vendorID)
	if err != nil {
		log.Printf("KDS: Error counting new orders: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	// Return a simple count or a refresh trigger if count > last_count
	// For simplicity, we just return the count as text
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%d", count)
}

func getVendorID(r *http.Request) string {
	user, ok := middlewares.GetUserFromContext(r.Context())
	if ok && user.Vendor != nil {
		return user.Vendor.ID
	}
	return "demo-vendor"
}
