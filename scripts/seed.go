package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	vendorID     = "demo-vendor"
	databaseName = "fube_local"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Load configuration
	cfg := config.Load()
	// Override database name if needed for seeding local env
	cfg.MongoDBName = databaseName

	fmt.Printf("Seeding MongoDB database: %s\n", cfg.MongoDBName)

	// Connect to MongoDB
	client, err := db.GetConnection(ctx, cfg)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to connect to MongoDB: %v", err)
	}
	database := client.Database(cfg.MongoDBName)

	// 1. Clear existing collections
	collections := []string{"materials", "yields", "menus"}
	for _, collName := range collections {
		fmt.Printf("Clearing collection: %s\n", collName)
		if err := database.Collection(collName).Drop(ctx); err != nil {
			log.Fatalf("CRITICAL: Failed to drop collection %s: %v", collName, err)
		}
	}

	// 2. Seed Materials
	materials := []models.MaterialDoc{
		{
			VendorID: vendorID,
			Name:     "Beef Patty",
			Price:    250.0, // in cents ($2.50)
			Unit:     "pcs",
		},
		{
			VendorID: vendorID,
			Name:     "Brioche Bun",
			Price:    80.0, // in cents ($0.80)
			Unit:     "pcs",
		},
		{
			VendorID: vendorID,
			Name:     "Tomato Sauce",
			Price:    500.0, // $5.00 per kg
			Unit:     "kg",
		},
		{
			VendorID: vendorID,
			Name:     "Lettuce",
			Price:    300.0, // $3.00 per kg
			Unit:     "kg",
		},
		{
			VendorID: vendorID,
			Name:     "Cheese",
			Price:    1000.0, // $10.00 per kg
			Unit:     "kg",
		},
	}

	materialMap := make(map[string]primitive.ObjectID)
	for i, mat := range materials {
		materials[i].CreatedAt = time.Now()
		materials[i].UpdatedAt = time.Now()
		res, err := database.Collection("materials").InsertOne(ctx, materials[i])
		if err != nil {
			log.Fatalf("CRITICAL: Failed to seed material %s: %v", mat.Name, err)
		}
		materialMap[mat.Name] = res.InsertedID.(primitive.ObjectID)
		fmt.Printf("Inserted material: %s (ID: %s)\n", mat.Name, materialMap[mat.Name].Hex())
	}

	// 3. Seed Yields
	// 'Burger Prep'
	burgerPrep := models.YieldDoc{
		VendorID:  vendorID,
		Name:      "Burger Prep",
		Unit:      "pcs",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Materials: []models.YieldMaterialDoc{
			{
				MaterialID: materialMap["Beef Patty"],
				Amount:     1,
				Unit:       "pcs",
			},
			{
				MaterialID: materialMap["Brioche Bun"],
				Amount:     1,
				Unit:       "pcs",
			},
			{
				MaterialID: materialMap["Tomato Sauce"],
				Amount:     0.02, // 20g
				Unit:       "kg",
			},
			{
				MaterialID: materialMap["Lettuce"],
				Amount:     0.015, // 15g
				Unit:       "kg",
			},
			{
				MaterialID: materialMap["Cheese"],
				Amount:     0.02, // 20g
				Unit:       "kg",
			},
		},
	}

	res, err := database.Collection("yields").InsertOne(ctx, burgerPrep)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to seed yield %s: %v", burgerPrep.Name, err)
	}
	burgerPrepID := res.InsertedID.(primitive.ObjectID)
	fmt.Printf("Inserted yield: %s (ID: %s)\n", burgerPrep.Name, burgerPrepID.Hex())

	// 4. Seed Menus
	menus := []models.MenuDoc{
		{
			VendorID:      vendorID,
			ExternalPosID: "POS-101",
			Name:          "Classic Burger",
			Price:         1200.0, // $12.00
			Unit:          "pcs",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Yields: []models.MenuYieldDoc{
				{
					YieldID: burgerPrepID,
					Amount:  1,
					Unit:    "pcs",
				},
			},
		},
		{
			VendorID:      vendorID,
			ExternalPosID: "POS-102",
			Name:          "Cheeseburger",
			Price:         1350.0, // $13.50
			Unit:          "pcs",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Yields: []models.MenuYieldDoc{
				{
					YieldID: burgerPrepID,
					Amount:  1,
					Unit:    "pcs",
				},
			},
		},
	}

	for _, menu := range menus {
		res, err := database.Collection("menus").InsertOne(ctx, menu)
		if err != nil {
			log.Fatalf("CRITICAL: Failed to seed menu %s: %v", menu.Name, err)
		}
		fmt.Printf("Inserted menu: %s (ID: %s, POS ID: %s)\n", menu.Name, res.InsertedID.(primitive.ObjectID).Hex(), menu.ExternalPosID)
	}

	fmt.Println("\nSuccessfully seeded local MongoDB database!")
	fmt.Println("To verify, you can use: mongosh \"mongodb://localhost:27017/fube_local\" --eval \"db.menus.find()\"")
}
