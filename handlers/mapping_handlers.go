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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetMappingDashboard renders the mapping dashboard showing all menu items.
// Tiger Style: Explicit error handling and clean template rendering.
func GetMappingDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middlewares.GetUserFromContext(ctx)
	vendorID := "demo-vendor" // Fallback for demo
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch all menu items for the vendor
	cursor, err := database.Collection("menus").Find(ctx, bson.M{"vendor_id": vendorID})
	if err != nil {
		log.Printf("Error fetching menus: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var menus []models.MenuDoc
	if err := cursor.All(ctx, &menus); err != nil {
		log.Printf("Error decoding menus: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/mapping_fragment.html",
	)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Menus []models.MenuDoc
	}{
		Menus: menus,
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// SearchYields returns an HTML fragment with yield search results.
// Tiger Style: Efficient MongoDB search with name regex.
func SearchYields(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")
	if query == "" {
		return
	}

	user, ok := middlewares.GetUserFromContext(ctx)
	vendorID := "demo-vendor"
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Search yields by name
	filter := bson.M{
		"vendor_id": vendorID,
		"name":      bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}},
	}
	findOptions := options.Find().SetLimit(10)
	cursor, err := database.Collection("yields").Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Error searching yields: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var yields []models.YieldDoc
	if err := cursor.All(ctx, &yields); err != nil {
		log.Printf("Error decoding yields: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/mapping_fragment.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "yield_list", yields)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// UpdateMenuMapping updates a MenuDoc with a selected YieldID and Amount.
// Tiger Style: Robust error handling and atomic-like update.
func UpdateMenuMapping(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	menuIDStr := r.FormValue("menu_id")
	yieldIDStr := r.FormValue("yield_id")
	amountStr := r.FormValue("amount")

	menuOID, err := primitive.ObjectIDFromHex(menuIDStr)
	if err != nil {
		http.Error(w, "Invalid menu ID", http.StatusBadRequest)
		return
	}

	yieldOID, err := primitive.ObjectIDFromHex(yieldIDStr)
	if err != nil {
		http.Error(w, "Invalid yield ID", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	user, ok := middlewares.GetUserFromContext(ctx)
	vendorID := "demo-vendor"
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch the yield to get its unit (optional, but good for consistency)
	var yield models.YieldDoc
	err = database.Collection("yields").FindOne(ctx, bson.M{"_id": yieldOID, "vendor_id": vendorID}).Decode(&yield)
	if err != nil {
		log.Printf("Yield not found: %v", err)
		http.Error(w, "Yield not found", http.StatusBadRequest)
		return
	}

	// Update the menu mapping
	// For simplicity, we'll replace the yields array with this single recipe mapping
	_, err = database.Collection("menus").UpdateOne(
		ctx,
		bson.M{"_id": menuOID, "vendor_id": vendorID},
		bson.M{
			"$set": bson.M{
				"yields": []models.MenuYieldDoc{
					{
						YieldID: yieldOID,
						Amount:  amount,
						Unit:    yield.Unit,
					},
				},
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		log.Printf("Error updating menu: %v", err)
		http.Error(w, "Failed to update menu mapping", http.StatusInternalServerError)
		return
	}

	// Re-fetch the menu to render the updated row
	var menu models.MenuDoc
	err = database.Collection("menus").FindOne(ctx, bson.M{"_id": menuOID, "vendor_id": vendorID}).Decode(&menu)
	if err != nil {
		http.Error(w, "Menu not found after update", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/mapping_fragment.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "menu_row", menu)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// RenderRecipePicker returns the HTML fragment for the recipe picker modal.
func RenderRecipePicker(w http.ResponseWriter, r *http.Request) {
	menuID := r.URL.Query().Get("menu_id")
	if menuID == "" {
		http.Error(w, "Missing menu_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, ok := middlewares.GetUserFromContext(ctx)
	vendorID := "demo-vendor"
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	menuOID, _ := primitive.ObjectIDFromHex(menuID)
	var menu models.MenuDoc
	err = database.Collection("menus").FindOne(ctx, bson.M{"_id": menuOID, "vendor_id": vendorID}).Decode(&menu)
	if err != nil {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/mapping_fragment.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		MenuID   string
		MenuName string
	}{
		MenuID:   menuID,
		MenuName: menu.Name,
	}

	err = tmpl.ExecuteTemplate(w, "recipe_picker", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
