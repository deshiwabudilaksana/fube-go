package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/gorilla/mux"
)

// GetMenuCostHandler handles the calculation of the total cost of a 'Menu' item.
// It uses MongoDB aggregation pipelines to join Menu -> Yield -> Material and calculate the total cost.
// Endpoint: GET /api/menus/{id}/cost
func GetMenuCostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	user, ok := middlewares.GetUserFromContext(r.Context())
	if !ok || user.Vendor == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		log.Printf("Error getting MongoDB database: %v", err)
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	vendorID := user.Vendor.ID
	result, err := models.CalculateMenuCostMongo(r.Context(), database, vendorID, idStr)
	if err != nil {
		log.Printf("Error getting menu cost for menu %s, vendor %s: %v", idStr, vendorID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PlanProductionHandler implements the 'Planning' function for raw materials.
// Input: List of {MenuID, PlannedQuantity}
// Output: Total aggregated amount of 'FoodMaterial' required across all menus.
// Endpoint: POST /api/planning/production
func PlanProductionHandler(w http.ResponseWriter, r *http.Request) {
	var input []models.ProductionPlanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body: expected JSON array of production plan inputs", http.StatusBadRequest)
		return
	}

	user, ok := middlewares.GetUserFromContext(r.Context())
	if !ok || user.Vendor == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		log.Printf("Error getting MongoDB database: %v", err)
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	vendorID := user.Vendor.ID
	result, err := models.GetProductionRequirementsMongo(r.Context(), database, vendorID, input)
	if err != nil {
		log.Printf("Error planning production for vendor %s: %v", vendorID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetExternalReportingHandler implements the 'External Reporting' function for all menus.
// It maps 'Menu.external_pos_id' to FUBE's calculated cost and yield data.
// Endpoint: GET /api/reports/external
func GetExternalReportingHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.GetUserFromContext(r.Context())
	if !ok || user.Vendor == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cfg := config.Load()
	database, err := db.GetDatabase(r.Context(), cfg)
	if err != nil {
		log.Printf("Error getting MongoDB database: %v", err)
		http.Error(w, "Database Connection Error", http.StatusInternalServerError)
		return
	}

	vendorID := user.Vendor.ID
	result, err := models.GetExternalReportingDataMongo(r.Context(), database, vendorID)
	if err != nil {
		log.Printf("Error getting external reporting data for vendor %s: %v", vendorID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
