package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/deshiwabudilaksana/fube-go/models"
	"go.mongodb.org/mongo-driver/mongo"
)

// ExporterService handles the generation of cost reports for external systems.
// Tiger Style: Explicit dependencies, robust error handling, and high-performance CSV encoding.
type ExporterService struct {
	db *mongo.Database
}

// NewExporterService creates a new instance of ExporterService.
func NewExporterService(db *mongo.Database) *ExporterService {
	return &ExporterService{db: db}
}

// GenerateCostReportCSV fetches menu items with external POS IDs and returns a CSV string.
// Headers: External ID, Name, Total Cost, Selling Price, Margin %
func (s *ExporterService) GenerateCostReportCSV(ctx context.Context, vendorID string) (string, error) {
	// Fetch reporting data using the optimized aggregation pipeline
	reportData, err := models.GetExternalReportingDataMongo(ctx, s.db, vendorID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch reporting data: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV headers
	headers := []string{"External ID", "Name", "Total Cost", "Selling Price", "Margin %"}
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write CSV headers: %w", err)
	}

	for _, data := range reportData {
		margin := 0.0
		if data.SellingPrice > 0 {
			margin = ((data.SellingPrice - data.TotalCost) / data.SellingPrice) * 100
		}

		row := []string{
			data.ExternalPosID,
			data.MenuName,
			fmt.Sprintf("%.2f", data.TotalCost),
			fmt.Sprintf("%.2f", data.SellingPrice),
			fmt.Sprintf("%.2f%%", margin),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row for menu %s: %w", data.MenuName, err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("csv writer flush error: %w", err)
	}

	return buf.String(), nil
}
