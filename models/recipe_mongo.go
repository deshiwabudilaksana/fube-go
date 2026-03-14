package models

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CalculateMenuCostMongo calculates the total cost of a menu item using MongoDB Aggregation Pipelines.
// It joins Menu -> Yield -> Material to find the price per ingredient and sums the total cost.
// Tiger Style: Explicit error handling, high performance through aggregation, and strict documentation.
func CalculateMenuCostMongo(ctx context.Context, db *mongo.Database, vendorID, menuID string) (*MenuCostResult, error) {
	menuOID, err := primitive.ObjectIDFromHex(menuID)
	if err != nil {
		return nil, fmt.Errorf("invalid menu ID format '%s': %w", menuID, err)
	}

	pipeline := mongo.Pipeline{
		// Match the specific menu and vendor
		{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: menuOID},
			{Key: "vendor_id", Value: vendorID},
		}}},
		// Unwind the yields array
		{{Key: "$unwind", Value: "$yields"}},
		// Lookup yield details
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "yields"},
			{Key: "localField", Value: "yields.yield_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "yield_data"},
		}}},
		{{Key: "$unwind", Value: "$yield_data"}},
		// Unwind the materials array within yield_data
		{{Key: "$unwind", Value: "$yield_data.materials"}},
		// Lookup material details
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "materials"},
			{Key: "localField", Value: "yield_data.materials.material_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "material_data"},
		}}},
		{{Key: "$unwind", Value: "$material_data"}},
		// Calculate cost per ingredient
		{{Key: "$addFields", Value: bson.D{
			{Key: "ingredient_cost", Value: bson.D{
				{Key: "$multiply", Value: bson.A{
					"$yields.amount",
					"$yield_data.materials.amount",
					"$material_data.price",
				}},
			}},
		}}},
		// Group by menu to sum total cost
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "menu_name", Value: bson.D{{Key: "$first", Value: "$name"}}},
			{Key: "total_cost", Value: bson.D{{Key: "$sum", Value: "$ingredient_cost"}}},
		}}},
	}

	cursor, err := db.Collection("menus").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute aggregate pipeline for menu cost: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID        primitive.ObjectID `bson:"_id"`
		MenuName  string             `bson:"menu_name"`
		TotalCost float64            `bson:"total_cost"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation results: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("menu not found or unauthorized: %s (vendor: %s)", menuID, vendorID)
	}

	return &MenuCostResult{
		MenuID:    results[0].ID.Hex(),
		MenuName:  results[0].MenuName,
		TotalCost: results[0].TotalCost,
	}, nil
}

// GetProductionRequirementsMongo aggregates multiple menus and their quantities to find the total sum of each material needed.
// Tiger Style: Robust error wrapping, efficient single database round-trip, and consistent naming.
func GetProductionRequirementsMongo(ctx context.Context, db *mongo.Database, vendorID string, plans []ProductionPlanInput) ([]MaterialRequirement, error) {
	if len(plans) == 0 {
		return []MaterialRequirement{}, nil
	}

	// Prepare mapping of menu IDs to their planned quantities for the pipeline
	menuIDs := make([]primitive.ObjectID, 0, len(plans))
	planMap := make(map[string]int)
	for _, p := range plans {
		oid, err := primitive.ObjectIDFromHex(p.MenuID)
		if err != nil {
			return nil, fmt.Errorf("invalid menu ID in production plan '%s': %w", p.MenuID, err)
		}
		menuIDs = append(menuIDs, oid)
		planMap[p.MenuID] += p.PlannedQuantity
	}

	// Construct the planned quantity lookup using $switch
	branches := make([]bson.D, 0, len(planMap))
	for mid, qty := range planMap {
		oid, _ := primitive.ObjectIDFromHex(mid)
		branches = append(branches, bson.D{
			{Key: "case", Value: bson.D{{Key: "$eq", Value: bson.A{"$_id", oid}}}},
			{Key: "then", Value: qty},
		})
	}

	pipeline := mongo.Pipeline{
		// Match menus in the plan
		{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$in", Value: menuIDs}}},
			{Key: "vendor_id", Value: vendorID},
		}}},
		// Assign planned quantity to each menu
		{{Key: "$addFields", Value: bson.D{
			{Key: "planned_quantity", Value: bson.D{
				{Key: "$switch", Value: bson.D{
					{Key: "branches", Value: branches},
					{Key: "default", Value: 0},
				}},
			}},
		}}},
		// Unwind yields and materials to calculate total requirements
		{{Key: "$unwind", Value: "$yields"}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "yields"},
			{Key: "localField", Value: "yields.yield_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "yield_data"},
		}}},
		{{Key: "$unwind", Value: "$yield_data"}},
		{{Key: "$unwind", Value: "$yield_data.materials"}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "materials"},
			{Key: "localField", Value: "yield_data.materials.material_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "material_data"},
		}}},
		{{Key: "$unwind", Value: "$material_data"}},
		// Calculate amount per material
		{{Key: "$addFields", Value: bson.D{
			{Key: "total_needed", Value: bson.D{
				{Key: "$multiply", Value: bson.A{
					"$planned_quantity",
					"$yields.amount",
					"$yield_data.materials.amount",
				}},
			}},
		}}},
		// Group by material ID to aggregate requirements
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$yield_data.materials.material_id"},
			{Key: "material_name", Value: bson.D{{Key: "$first", Value: "$material_data.name"}}},
			{Key: "total_amount", Value: bson.D{{Key: "$sum", Value: "$total_needed"}}},
			{Key: "unit", Value: bson.D{{Key: "$first", Value: "$material_data.unit"}}},
		}}},
		// Sort by material name for consistency
		{{Key: "$sort", Value: bson.D{{Key: "material_name", Value: 1}}}},
	}

	cursor, err := db.Collection("menus").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate production requirements: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		MaterialID   primitive.ObjectID `bson:"_id"`
		MaterialName string             `bson:"material_name"`
		TotalAmount  float64            `bson:"total_amount"`
		Unit         string             `bson:"unit"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode production requirement results: %w", err)
	}

	requirements := make([]MaterialRequirement, 0, len(results))
	for _, r := range results {
		requirements = append(requirements, MaterialRequirement{
			MaterialID:   r.MaterialID.Hex(),
			MaterialName: r.MaterialName,
			TotalAmount:  r.TotalAmount,
			Unit:         r.Unit,
		})
	}

	return requirements, nil
}

// GetExternalReportingDataMongo generates a JSON payload mapping Menu.external_pos_id to FUBE's calculated cost/yield data.
// It performs a complex aggregation to calculate total cost and collect yield details in a single database round-trip.
func GetExternalReportingDataMongo(ctx context.Context, db *mongo.Database, vendorID string) ([]ExternalReportData, error) {
	pipeline := mongo.Pipeline{
		// Match menus with an external POS ID for this vendor
		{{Key: "$match", Value: bson.D{
			{Key: "external_pos_id", Value: bson.D{{Key: "$ne", Value: nil}}},
			{Key: "vendor_id", Value: vendorID},
		}}},
		// Unwind yields to join with yield and material data
		{{Key: "$unwind", Value: "$yields"}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "yields"},
			{Key: "localField", Value: "yields.yield_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "yield_info"},
		}}},
		{{Key: "$unwind", Value: "$yield_info"}},
		{{Key: "$unwind", Value: "$yield_info.materials"}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "materials"},
			{Key: "localField", Value: "yield_info.materials.material_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "material_info"},
		}}},
		{{Key: "$unwind", Value: "$material_info"}},
		// Calculate cost contribution of each material
		{{Key: "$addFields", Value: bson.D{
			{Key: "material_cost", Value: bson.D{
				{Key: "$multiply", Value: bson.A{
					"$yields.amount",
					"$yield_info.materials.amount",
					"$material_info.price",
				}},
			}},
		}}},
		// Group by Menu and Yield to get cost per yield and avoid duplicate yields in final list
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "menu_id", Value: "$_id"},
				{Key: "yield_id", Value: "$yield_info._id"},
			}},
			{Key: "menu_name", Value: bson.D{{Key: "$first", Value: "$name"}}},
			{Key: "external_pos_id", Value: bson.D{{Key: "$first", Value: "$external_pos_id"}}},
			{Key: "selling_price", Value: bson.D{{Key: "$first", Value: "$price"}}},
			{Key: "yield_name", Value: bson.D{{Key: "$first", Value: "$yield_info.name"}}},
			{Key: "yield_amount", Value: bson.D{{Key: "$first", Value: "$yields.amount"}}},
			{Key: "yield_unit", Value: bson.D{{Key: "$first", Value: "$yields.unit"}}},
			{Key: "yield_cost", Value: bson.D{{Key: "$sum", Value: "$material_cost"}}},
		}}},
		// Group by Menu to aggregate all yields and sum total cost
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id.menu_id"},
			{Key: "menu_name", Value: bson.D{{Key: "$first", Value: "$menu_name"}}},
			{Key: "external_pos_id", Value: bson.D{{Key: "$first", Value: "$external_pos_id"}}},
			{Key: "selling_price", Value: bson.D{{Key: "$first", Value: "$selling_price"}}},
			{Key: "total_cost", Value: bson.D{{Key: "$sum", Value: "$yield_cost"}}},
			{Key: "yields", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "name", Value: "$yield_name"},
					{Key: "amount", Value: "$yield_amount"},
					{Key: "unit", Value: "$yield_unit"},
				}},
			}},
		}}},
	}

	cursor, err := db.Collection("menus").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate external reporting data: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ExternalPosID string  `bson:"external_pos_id"`
		MenuName      string  `bson:"menu_name"`
		TotalCost     float64 `bson:"total_cost"`
		SellingPrice  float64 `bson:"selling_price"`
		Yields        []struct {
			Name   string  `bson:"name"`
			Amount float64 `bson:"amount"`
			Unit   string  `bson:"unit"`
		} `bson:"yields"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode reporting results: %w", err)
	}

	report := make([]ExternalReportData, 0, len(results))
	for _, r := range results {
		data := ExternalReportData{
			ExternalPosID: r.ExternalPosID,
			MenuName:      r.MenuName,
			TotalCost:     r.TotalCost,
			SellingPrice:  r.SellingPrice,
		}
		for _, y := range r.Yields {
			data.Yields = append(data.Yields, struct {
				Name   string `json:"name"`
				Amount int    `json:"amount"`
				Unit   string `json:"unit"`
			}{
				Name:   y.Name,
				Amount: int(y.Amount), // Convert to int as per existing struct requirement
				Unit:   y.Unit,
			})
		}
		report = append(report, data)
	}

	return report, nil
}
