package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/models"
	"go.mongodb.org/mongo-driver/bson"
)

// RenderLandingPage renders the public landing page for user onboarding
func RenderLandingPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/landing.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}

// MenuItem represents a menu item in the yield planning table
type MenuItem struct {
	ID         string
	Name       string
	UnitCost   float64
	PlannedQty int
	TotalCost  float64
}

// YieldPlanningData is the data structure for the yield planning template
type YieldPlanningData struct {
	MenuItems         []MenuItem
	RequiredMaterials []models.MaterialRequirement
	GrandTotal        float64
}

// getYieldData fetches real data from MongoDB based on current plans
func getYieldData(ctx context.Context, vendorID string, plans []models.ProductionPlanInput) (YieldPlanningData, error) {
	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		return YieldPlanningData{}, err
	}

	// 1. Fetch all menus for this vendor
	cursor, err := database.Collection("menus").Find(ctx, bson.M{"vendor_id": vendorID})
	if err != nil {
		return YieldPlanningData{}, err
	}
	defer cursor.Close(ctx)

	var menuDocs []models.MenuDoc
	if err := cursor.All(ctx, &menuDocs); err != nil {
		return YieldPlanningData{}, err
	}

	// 2. Process menu items and calculate costs
	var menuItems []MenuItem
	var grandTotal float64
	for _, doc := range menuDocs {
		qty := 0
		for _, p := range plans {
			if p.MenuID == doc.ID.Hex() {
				qty = p.PlannedQuantity
				break
			}
		}

		// Calculate unit cost using existing aggregation logic
		costResult, err := models.CalculateMenuCostMongo(ctx, database, vendorID, doc.ID.Hex())
		unitCost := 0.0
		if err == nil && costResult != nil {
			unitCost = costResult.TotalCost
		}

		totalCost := unitCost * float64(qty)
		menuItems = append(menuItems, MenuItem{
			ID:         doc.ID.Hex(),
			Name:       doc.Name,
			UnitCost:   unitCost,
			PlannedQty: qty,
			TotalCost:  totalCost,
		})
		grandTotal += totalCost
	}

	// 3. Get material requirements using aggregation logic
	requirements, err := models.GetProductionRequirementsMongo(ctx, database, vendorID, plans)
	if err != nil {
		return YieldPlanningData{}, err
	}

	return YieldPlanningData{
		MenuItems:         menuItems,
		RequiredMaterials: requirements,
		GrandTotal:        grandTotal,
	}, nil
}

// RenderYieldPlanning renders the full page with the layout
func RenderYieldPlanning(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.GetUserFromContext(r.Context())
	vendorID := "demo-vendor" // Fallback for demo
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	data, err := getYieldData(r.Context(), vendorID, []models.ProductionPlanInput{})
	if err != nil {
		log.Printf("Error fetching yield data: %v", err)
		// Still try to render with empty data or show error
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/yield_planning_fragment.html",
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}

// RenderImport renders the CSV import page
func RenderImport(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/import_fragment.html",
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}

// UpdateYieldPlanning handles HTMX requests for updating quantities
func UpdateYieldPlanning(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	user, ok := middlewares.GetUserFromContext(r.Context())
	vendorID := "demo-vendor" // Fallback for demo
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	// Extract all quantities from the form
	var plans []models.ProductionPlanInput
	for key, values := range r.Form {
		if strings.HasPrefix(key, "qty-") && len(values) > 0 {
			menuID := strings.TrimPrefix(key, "qty-")
			qty, _ := strconv.Atoi(values[0])
			if qty > 0 {
				plans = append(plans, models.ProductionPlanInput{
					MenuID:          menuID,
					PlannedQuantity: qty,
				})
			}
		}
	}

	// Also check for 'menu_id' and 'quantity' if sent individually by HTMX
	if mid := r.FormValue("menu_id"); mid != "" {
		qty, _ := strconv.Atoi(r.FormValue("quantity"))
		found := false
		for i, p := range plans {
			if p.MenuID == mid {
				plans[i].PlannedQuantity = qty
				found = true
				break
			}
		}
		if !found && qty > 0 {
			plans = append(plans, models.ProductionPlanInput{
				MenuID:          mid,
				PlannedQuantity: qty,
			})
		}
	}

	data, err := getYieldData(r.Context(), vendorID, plans)
	if err != nil {
		log.Printf("Error updating yield data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/yield_planning_fragment.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	// Execute only the "main" block for HTMX partial update
	err = tmpl.ExecuteTemplate(w, "main", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}
