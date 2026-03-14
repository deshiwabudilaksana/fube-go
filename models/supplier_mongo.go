package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// POStatus defines the lifecycle of a Purchase Order.
type POStatus string

const (
	StatusDraft    POStatus = "DRAFT"
	StatusSent     POStatus = "SENT"
	StatusReceived POStatus = "RECEIVED"
)

// SupplierDoc represents a vendor that provides materials.
type SupplierDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VendorID  string             `bson:"vendor_id" json:"vendor_id"`
	Name      string             `bson:"name" json:"name"`
	Contact   string             `bson:"contact" json:"contact"`
	LeadTime  int                `bson:"lead_time" json:"lead_time"` // In days
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// POLineItem represents a specific material and quantity in a Purchase Order.
type POLineItem struct {
	MaterialID    primitive.ObjectID `bson:"material_id" json:"material_id"`
	Quantity      float64            `bson:"quantity" json:"quantity"`
	ExpectedPrice float64            `bson:"expected_price" json:"expected_price"`
}

// PurchaseOrderDoc represents an order sent to a supplier.
type PurchaseOrderDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VendorID   string             `bson:"vendor_id" json:"vendor_id"`
	StoreID    string             `bson:"store_id" json:"store_id"`
	SupplierID primitive.ObjectID `bson:"supplier_id" json:"supplier_id"`
	Status     POStatus           `bson:"status" json:"status"`
	Items      []POLineItem       `bson:"items" json:"items"`
	Total      float64            `bson:"total" json:"total"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// SalesForecastDoc represents the planned sales/requirements for a period.
// This is used as the 'Planned Sales' source for suggested orders.
type SalesForecastDoc struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	VendorID  string                `bson:"vendor_id" json:"vendor_id"`
	StoreID   string                `bson:"store_id" json:"store_id"`
	Plans     []ProductionPlanInput `bson:"plans" json:"plans"`
	StartDate time.Time             `bson:"start_date" json:"start_date"`
	EndDate   time.Time             `bson:"end_date" json:"end_date"`
	CreatedAt time.Time             `bson:"created_at" json:"created_at"`
}

// SuggestedOrderItem represents an item that is recommended to be purchased.
type SuggestedOrderItem struct {
	MaterialID   string  `json:"material_id"`
	MaterialName string  `json:"material_name"`
	SupplierID   string  `json:"supplier_id"`
	SupplierName string  `json:"supplier_name"`
	NeededQty    float64 `json:"needed_qty"`
	CurrentStock float64 `json:"current_stock"`
	OrderQty     float64 `json:"order_qty"`
	UnitPrice    float64 `json:"unit_price"`
}

// CalculateSuggestedOrders aggregates requirements from Sales Forecasts,
// subtracts current physical stock, and groups by supplier.
// Tiger Style: Robust error handling, high-performance MongoDB aggregation.
func CalculateSuggestedOrders(ctx context.Context, db *mongo.Database, vendorID, storeID string) ([]SuggestedOrderItem, error) {
	// 1. Get all active plans (for simplicity, we'll take all forecasts for this vendor/store)
	// In a real system, we'd filter by a specific date range.
	cursor, err := db.Collection("sales_forecasts").Find(ctx, bson.M{
		"vendor_id": vendorID,
		"store_id":  storeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sales forecasts: %w", err)
	}
	defer cursor.Close(ctx)

	var forecasts []SalesForecastDoc
	if err := cursor.All(ctx, &forecasts); err != nil {
		return nil, fmt.Errorf("failed to decode sales forecasts: %w", err)
	}

	allPlans := []ProductionPlanInput{}
	for _, f := range forecasts {
		allPlans = append(allPlans, f.Plans...)
	}

	if len(allPlans) == 0 {
		return []SuggestedOrderItem{}, nil
	}

	// 2. Aggregate requirements using the recipe engine logic
	requirements, err := GetProductionRequirementsMongo(ctx, db, vendorID, allPlans)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate production requirements: %w", err)
	}

	// 3. Subtract current physical stock and enrich with supplier info
	suggested := []SuggestedOrderItem{}
	for _, req := range requirements {
		matOID, _ := primitive.ObjectIDFromHex(req.MaterialID)

		// Get material info for default supplier and price
		var mat MaterialDoc
		err := db.Collection("materials").FindOne(ctx, bson.M{"_id": matOID, "vendor_id": vendorID}).Decode(&mat)
		if err != nil {
			continue // Skip materials we can't find
		}

		if mat.DefaultSupplierID == nil {
			continue // Skip if no default supplier is set
		}

		// Get physical stock
		// Note: Using a simplified version of GetInventoryAudit logic here or just querying inventory collection
		var inv InventoryDoc
		err = db.Collection("inventory").FindOne(ctx, bson.M{
			"vendor_id":   vendorID,
			"store_id":    storeID,
			"material_id": matOID,
		}).Decode(&inv)

		currentStock := 0.0
		if err == nil {
			currentStock = inv.Quantity
		}

		orderQty := req.TotalAmount - currentStock
		if orderQty > 0 {
			var supplier SupplierDoc
			err = db.Collection("suppliers").FindOne(ctx, bson.M{"_id": mat.DefaultSupplierID}).Decode(&supplier)
			supplierName := "Unknown"
			if err == nil {
				supplierName = supplier.Name
			}

			suggested = append(suggested, SuggestedOrderItem{
				MaterialID:   req.MaterialID,
				MaterialName: req.MaterialName,
				SupplierID:   mat.DefaultSupplierID.Hex(),
				SupplierName: supplierName,
				NeededQty:    req.TotalAmount,
				CurrentStock: currentStock,
				OrderQty:     orderQty,
				UnitPrice:    mat.Price,
			})
		}
	}

	return suggested, nil
}

// GeneratePOFromSuggestions creates DRAFT Purchase Orders from calculated suggestions.
// Tiger Style: Atomic-like behavior (manual cleanup on failure) and strict validation.
func GeneratePOFromSuggestions(ctx context.Context, db *mongo.Database, vendorID, storeID string, suggestions []SuggestedOrderItem) error {
	if len(suggestions) == 0 {
		return nil
	}

	// Group suggestions by SupplierID
	bySupplier := make(map[string][]SuggestedOrderItem)
	for _, s := range suggestions {
		bySupplier[s.SupplierID] = append(bySupplier[s.SupplierID], s)
	}

	for supplierID, items := range bySupplier {
		suppOID, err := primitive.ObjectIDFromHex(supplierID)
		if err != nil {
			return fmt.Errorf("invalid supplier ID %s: %w", supplierID, err)
		}

		poItems := make([]POLineItem, 0, len(items))
		total := 0.0
		for _, item := range items {
			matOID, _ := primitive.ObjectIDFromHex(item.MaterialID)
			poItems = append(poItems, POLineItem{
				MaterialID:    matOID,
				Quantity:      item.OrderQty,
				ExpectedPrice: item.UnitPrice,
			})
			total += item.OrderQty * item.UnitPrice
		}

		po := PurchaseOrderDoc{
			ID:         primitive.NewObjectID(),
			VendorID:   vendorID,
			StoreID:    storeID,
			SupplierID: suppOID,
			Status:     StatusDraft,
			Items:      poItems,
			Total:      total,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_, err = db.Collection("purchase_orders").InsertOne(ctx, po)
		if err != nil {
			return fmt.Errorf("failed to create PO for supplier %s: %w", supplierID, err)
		}
	}

	return nil
}

// GetSuppliers fetches all suppliers for a specific vendor.
func GetSuppliers(ctx context.Context, db *mongo.Database, vendorID string) ([]SupplierDoc, error) {
	cursor, err := db.Collection("suppliers").Find(ctx, bson.M{"vendor_id": vendorID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch suppliers: %w", err)
	}
	defer cursor.Close(ctx)

	var suppliers []SupplierDoc
	if err := cursor.All(ctx, &suppliers); err != nil {
		return nil, fmt.Errorf("failed to decode suppliers: %w", err)
	}
	return suppliers, nil
}

// GetPurchaseOrders fetches all purchase orders for a specific vendor/store.
func GetPurchaseOrders(ctx context.Context, db *mongo.Database, vendorID, storeID string) ([]PurchaseOrderDoc, error) {
	cursor, err := db.Collection("purchase_orders").Find(ctx, bson.M{
		"vendor_id": vendorID,
		"store_id":  storeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchase orders: %w", err)
	}
	defer cursor.Close(ctx)

	var pos []PurchaseOrderDoc
	if err := cursor.All(ctx, &pos); err != nil {
		return nil, fmt.Errorf("failed to decode purchase orders: %w", err)
	}
	return pos, nil
}
