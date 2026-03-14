package importer

import (
	"context"
	"strings"
	"testing"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestImportCSV(t *testing.T) {
	// Setup test database configuration
	ctx := context.Background()
	cfg := &config.Config{
		MongoURI:    "mongodb://localhost:27017",
		MongoDBName: "fube_importer_test",
	}

	// Connect to verify and setup cleanup
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		t.Skip("Skipping MongoDB-dependent test: local MongoDB not available")
		return
	}
	defer client.Disconnect(ctx)

	testDB := client.Database(cfg.MongoDBName)
	// Ensure we start fresh
	_ = testDB.Drop(ctx)
	defer testDB.Drop(ctx)

	service := NewImporterService(cfg)

	tests := []struct {
		name           string
		csvData        string
		vendorID       string
		storeID        string
		expectedMenus  int
		expectedMats   int
		validateResult func(t *testing.T, db *mongo.Database)
	}{
		{
			name: "Square Style Menu Items",
			csvData: "Item,Price,SKU,Unit\n" +
				"Classic Burger,12.50,BRG-01,ea\n" +
				"Cheese Pizza,15.00,PZ-02,ea",
			vendorID:      "v1",
			storeID:       "s1",
			expectedMenus: 2,
			expectedMats:  0,
			validateResult: func(t *testing.T, db *mongo.Database) {
				var menus []models.MenuDoc
				cursor, err := db.Collection("menu").Find(ctx, bson.M{"vendor_id": "v1"})
				require.NoError(t, err)
				err = cursor.All(ctx, &menus)
				require.NoError(t, err)
				assert.Equal(t, "Classic Burger", menus[0].Name)
				assert.Equal(t, 12.50, menus[0].Price)
				assert.Equal(t, "BRG-01", menus[0].ExternalPosID)
			},
		},
		{
			name: "Material Items Detection",
			csvData: "Material,Cost,Unit,Category\n" +
				"Tomato Sauce,2.50,kg,Ingredients material\n" +
				"Dough Ball,1.00,unit,Dough material",
			vendorID:      "v2",
			storeID:       "s2",
			expectedMenus: 0,
			expectedMats:  2,
			validateResult: func(t *testing.T, db *mongo.Database) {
				var mats []models.MaterialDoc
				cursor, err := db.Collection("materials").Find(ctx, bson.M{"vendor_id": "v2"})
				require.NoError(t, err)
				err = cursor.All(ctx, &mats)
				require.NoError(t, err)
				assert.Equal(t, "Tomato Sauce", mats[0].Name)
				assert.Equal(t, 2.50, mats[0].Price)
				assert.Equal(t, "kg", mats[0].Unit)
			},
		},
		{
			name: "Toast Style Headers",
			csvData: "Name,Price,ID,Measure\n" +
				"Iced Coffee,4.00,EXT-99,cups",
			vendorID:      "v3",
			storeID:       "s3",
			expectedMenus: 1,
			expectedMats:  0,
			validateResult: func(t *testing.T, db *mongo.Database) {
				var menu models.MenuDoc
				err := db.Collection("menu").FindOne(ctx, bson.M{"external_pos_id": "EXT-99"}).Decode(&menu)
				require.NoError(t, err)
				assert.Equal(t, "Iced Coffee", menu.Name)
				assert.Equal(t, "cups", menu.Unit)
			},
		},
		{
			name: "Malformed Rows and Missing Columns",
			csvData: "Item,Price\n" +
				"Valid Item,10.00\n" +
				"Invalid Row\n" +
				",20.00\n" + // Missing name - should be skipped
				"No Price Item,",
			vendorID:      "v4",
			storeID:       "s4",
			expectedMenus: 3, // "Valid Item", "Invalid Row", and "No Price Item"
			expectedMats:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.csvData)
			err := service.ImportCSV(ctx, tt.vendorID, tt.storeID, r)
			require.NoError(t, err)

			menuCount, _ := testDB.Collection("menu").CountDocuments(ctx, bson.M{"vendor_id": tt.vendorID})
			matCount, _ := testDB.Collection("materials").CountDocuments(ctx, bson.M{"vendor_id": tt.vendorID})

			assert.Equal(t, int64(tt.expectedMenus), menuCount, "Menu count mismatch")
			assert.Equal(t, int64(tt.expectedMats), matCount, "Material count mismatch")

			if tt.validateResult != nil {
				tt.validateResult(t, testDB)
			}
		})
	}
}
