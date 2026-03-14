package models

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testDB *mongo.Database

func TestMain(m *testing.M) {
	// Setup: Connect to local MongoDB for tests
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		// If MongoDB is not running, we might want to skip these tests or fail.
		// For a CI environment, MongoDB should be available.
		panic(err)
	}

	testDB = client.Database("fube_test")

	// Run tests
	code := m.Run()

	// Teardown: Clean up test database
	_ = testDB.Drop(ctx)
	_ = client.Disconnect(ctx)

	os.Exit(code)
}

func setupCollection(t *testing.T, collName string) *mongo.Collection {
	coll := testDB.Collection(collName)
	err := coll.Drop(context.Background())
	require.NoError(t, err)
	return coll
}

func TestCalculateMenuCostMongo(t *testing.T) {
	ctx := context.Background()
	vendorID := "vendor_123"

	// 1. Seed Materials
	matColl := setupCollection(t, "materials")
	matID1 := primitive.NewObjectID()
	matID2 := primitive.NewObjectID()
	_, err := matColl.InsertMany(ctx, []interface{}{
		MaterialDoc{
			ID:       matID1,
			VendorID: vendorID,
			Name:     "Flour",
			Price:    2.0,
			Unit:     "kg",
		},
		MaterialDoc{
			ID:       matID2,
			VendorID: vendorID,
			Name:     "Sugar",
			Price:    1.5,
			Unit:     "kg",
		},
	})
	require.NoError(t, err)

	// 2. Seed Yields (Recipes)
	yieldColl := setupCollection(t, "yields")
	yieldID1 := primitive.NewObjectID()
	_, err = yieldColl.InsertOne(ctx, YieldDoc{
		ID:       yieldID1,
		VendorID: vendorID,
		Name:     "Dough",
		Materials: []YieldMaterialDoc{
			{MaterialID: matID1, Amount: 0.5}, // 0.5kg * $2 = $1.0
			{MaterialID: matID2, Amount: 0.2}, // 0.2kg * $1.5 = $0.3
		},
	})
	require.NoError(t, err)

	// 3. Seed Menus
	menuColl := setupCollection(t, "menus")
	menuID := primitive.NewObjectID()
	_, err = menuColl.InsertOne(ctx, MenuDoc{
		ID:       menuID,
		VendorID: vendorID,
		Name:     "Bread",
		Yields: []MenuYieldDoc{
			{YieldID: yieldID1, Amount: 2.0}, // 2.0 * ($1.0 + $0.3) = $2.6
		},
	})
	require.NoError(t, err)

	t.Run("Successful Calculation", func(t *testing.T) {
		result, err := CalculateMenuCostMongo(ctx, testDB, vendorID, menuID.Hex())
		require.NoError(t, err)
		assert.Equal(t, menuID.Hex(), result.MenuID)
		assert.Equal(t, "Bread", result.MenuName)
		assert.InDelta(t, 2.6, result.TotalCost, 0.001)
	})

	t.Run("Menu Not Found", func(t *testing.T) {
		_, err := CalculateMenuCostMongo(ctx, testDB, vendorID, primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "menu not found")
	})

	t.Run("Unauthorized Vendor", func(t *testing.T) {
		_, err := CalculateMenuCostMongo(ctx, testDB, "wrong_vendor", menuID.Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "menu not found or unauthorized")
	})

	t.Run("Invalid Menu ID Format", func(t *testing.T) {
		_, err := CalculateMenuCostMongo(ctx, testDB, vendorID, "invalid-hex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid menu ID format")
	})
}

func TestGetProductionRequirementsMongo(t *testing.T) {
	ctx := context.Background()
	vendorID := "vendor_456"

	// 1. Seed Materials
	matColl := setupCollection(t, "materials")
	matID1 := primitive.NewObjectID()
	matID2 := primitive.NewObjectID()
	_, err := matColl.InsertMany(ctx, []interface{}{
		MaterialDoc{ID: matID1, VendorID: vendorID, Name: "Material A", Price: 1.0, Unit: "oz"},
		MaterialDoc{ID: matID2, VendorID: vendorID, Name: "Material B", Price: 1.0, Unit: "oz"},
	})
	require.NoError(t, err)

	// 2. Seed Yields
	yieldColl := setupCollection(t, "yields")
	yieldID1 := primitive.NewObjectID()
	_, err = yieldColl.InsertOne(ctx, YieldDoc{
		ID:       yieldID1,
		VendorID: vendorID,
		Name:     "Base Recipe",
		Materials: []YieldMaterialDoc{
			{MaterialID: matID1, Amount: 10},
			{MaterialID: matID2, Amount: 5},
		},
	})
	require.NoError(t, err)

	// 3. Seed Menus
	menuColl := setupCollection(t, "menus")
	menuID1 := primitive.NewObjectID()
	menuID2 := primitive.NewObjectID()
	_, err = menuColl.InsertMany(ctx, []interface{}{
		MenuDoc{
			ID:       menuID1,
			VendorID: vendorID,
			Name:     "Product 1",
			Yields: []MenuYieldDoc{
				{YieldID: yieldID1, Amount: 1.0},
			},
		},
		MenuDoc{
			ID:       menuID2,
			VendorID: vendorID,
			Name:     "Product 2",
			Yields: []MenuYieldDoc{
				{YieldID: yieldID1, Amount: 2.0},
			},
		},
	})
	require.NoError(t, err)

	t.Run("Combined Production Plan", func(t *testing.T) {
		plans := []ProductionPlanInput{
			{MenuID: menuID1.Hex(), PlannedQuantity: 10}, // Menu 1: 10 * (1.0 * Yield 1). Mat A: 10*10=100, Mat B: 10*5=50
			{MenuID: menuID2.Hex(), PlannedQuantity: 5},  // Menu 2: 5 * (2.0 * Yield 1). Mat A: 10*10=100, Mat B: 10*5=50
		}
		// Expected Total: Mat A: 200, Mat B: 100

		results, err := GetProductionRequirementsMongo(ctx, testDB, vendorID, plans)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// Results should be sorted by MaterialName (A then B)
		assert.Equal(t, "Material A", results[0].MaterialName)
		assert.Equal(t, 200.0, results[0].TotalAmount)
		assert.Equal(t, "Material B", results[1].MaterialName)
		assert.Equal(t, 100.0, results[1].TotalAmount)
	})

	t.Run("Empty Production Plan", func(t *testing.T) {
		results, err := GetProductionRequirementsMongo(ctx, testDB, vendorID, []ProductionPlanInput{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})

	t.Run("Invalid Menu ID in Plan", func(t *testing.T) {
		plans := []ProductionPlanInput{
			{MenuID: "invalid-id", PlannedQuantity: 1},
		}
		_, err := GetProductionRequirementsMongo(ctx, testDB, vendorID, plans)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid menu ID")
	})
}
