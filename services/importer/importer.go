package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ImporterService handles CSV data ingestion into MongoDB.
// Tiger Style: Explicit error wrapping, context-aware operations, and robust mapping.
type ImporterService struct {
	cfg *config.Config
}

// NewImporterService initializes a new ImporterService.
func NewImporterService(cfg *config.Config) *ImporterService {
	return &ImporterService{cfg: cfg}
}

// ImportCSV parses the CSV data and maps it to MenuDoc and MaterialDoc models.
// It uses VendorID and StoreID to scope the data.
func (s *ImporterService) ImportCSV(ctx context.Context, vendorID, storeID string, r io.Reader) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // Allow records with different number of fields
	// Read header
	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map header indices
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[strings.ToLower(strings.TrimSpace(name))] = i
	}

	mongoDB, err := db.GetDatabase(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("failed to get MongoDB database: %w", err)
	}

	menuColl := mongoDB.Collection("menu")
	materialColl := mongoDB.Collection("materials")

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record: %w", err)
		}

		// Try to map to MenuDoc or MaterialDoc
		// We use heuristics or specific columns for differentiation.
		// Square/Toast usually have "Item" or "Name".
		name := getValue(record, colMap, "item", "name", "material")
		if name == "" {
			continue // Skip rows without a name
		}

		priceStr := getValue(record, colMap, "price", "cost")
		price, _ := strconv.ParseFloat(priceStr, 64)

		externalID := getValue(record, colMap, "external_id", "sku", "id", "token")
		unit := getValue(record, colMap, "unit", "measure")
		if unit == "" {
			unit = "unit" // Default unit
		}

		// Heuristic: if it's in a column specifically named "material" or if "price" is 0 but "cost" is > 0,
		// or if the category implies it's an ingredient.
		// For simplicity, we'll import as MenuDoc if ExternalPosID is present, otherwise MaterialDoc.
		// Or better: check if "price" or "cost" column is used.

		isMaterial := strings.Contains(strings.ToLower(getValue(record, colMap, "category", "type")), "material") ||
			getValue(record, colMap, "material") != ""

		if isMaterial {
			material := models.MaterialDoc{
				ID:        primitive.NewObjectID(),
				VendorID:  vendorID,
				StoreID:   storeID,
				Name:      name,
				Price:     price,
				Unit:      unit,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_, err = materialColl.InsertOne(ctx, material)
			if err != nil {
				return fmt.Errorf("failed to insert MaterialDoc (%s): %w", name, err)
			}
		} else {
			menu := models.MenuDoc{
				ID:            primitive.NewObjectID(),
				VendorID:      vendorID,
				StoreID:       storeID,
				ExternalPosID: externalID,
				Name:          name,
				Price:         price,
				Unit:          unit,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			_, err = menuColl.InsertOne(ctx, menu)
			if err != nil {
				return fmt.Errorf("failed to insert MenuDoc (%s): %w", name, err)
			}
		}
	}

	return nil
}

func getValue(record []string, colMap map[string]int, keys ...string) string {
	for _, k := range keys {
		if idx, ok := colMap[k]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
	}
	return ""
}
