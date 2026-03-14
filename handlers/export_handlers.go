package handlers

import (
	"log"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/middlewares"
	"github.com/deshiwabudilaksana/fube-go/services/export"
)

// DownloadCostReportHandler handles the generation and download of the cost report CSV.
// Tiger Style: Proper content headers, robust error handling, and scoped to the user's vendor.
func DownloadCostReportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middlewares.GetUserFromContext(ctx)
	vendorID := "demo-vendor" // Fallback for demo
	if ok && user.Vendor != nil {
		vendorID = user.Vendor.ID
	}

	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		log.Printf("DownloadCostReportHandler: Error getting database: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	exporter := export.NewExporterService(database)
	csvData, err := exporter.GenerateCostReportCSV(ctx, vendorID)
	if err != nil {
		log.Printf("DownloadCostReportHandler: Failed to generate CSV for vendor %s: %v", vendorID, err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=fube_cost_report.csv")

	if _, err := w.Write([]byte(csvData)); err != nil {
		log.Printf("DownloadCostReportHandler: Error writing CSV response: %v", err)
	}
}
