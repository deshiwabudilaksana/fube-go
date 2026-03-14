package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/services/importer"
)

// ImportCSVHandler handles the CSV file upload and processing.
// It supports multi-part form upload and scopes imports to the authenticated user's VendorID and StoreID.
// Tiger Style: Robust error handling, proper resource management, and HTMX-compatible responses.
func ImportCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	user, ok := middlewares.GetUserFromContext(r.Context())
	if !ok || user.Vendor == nil {
		log.Println("ImportCSVHandler: Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Limit upload size to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("ImportCSVHandler: Failed to parse multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		log.Printf("ImportCSVHandler: No csv_file in form: %v", err)
		http.Error(w, "File 'csv_file' is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	vendorID := user.Vendor.ID
	// Get StoreID from form, default to 'main' if not provided.
	// In a full implementation, StoreID might come from user context if they are scoped to a specific store.
	storeID := r.FormValue("store_id")
	if storeID == "" {
		storeID = "main"
	}

	cfg := config.Load()
	svc := importer.NewImporterService(cfg)

	// Process CSV
	if err := svc.ImportCSV(r.Context(), vendorID, storeID, file); err != nil {
		log.Printf("ImportCSVHandler: CSV processing failed for vendor %s: %v", vendorID, err)
		w.WriteHeader(http.StatusInternalServerError)
		// Return HTML fragment for HTMX to display
		fmt.Fprintf(w, `<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
			<strong class="font-bold">Import failed!</strong>
			<span class="block sm:inline">%v</span>
		</div>`, err)
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative" role="alert">
		<strong class="font-bold">Success!</strong>
		<span class="block sm:inline">CSV data has been successfully imported and linked to your items.</span>
	</div>`)
}
