package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetKDSOrders fetches all active orders for the kitchen display.
// Tiger Style: Strict status filtering and vendor scoping.
func GetKDSOrders(ctx context.Context, db *mongo.Database, vendorID string) ([]OrderDoc, error) {
	filter := bson.M{
		"vendor_id": vendorID,
		"status":    bson.M{"$in": []OrderStatus{StatusPending, StatusConfirmed, StatusPreparing}},
	}

	cursor, err := db.Collection("orders").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch KDS orders for vendor %s: %w", vendorID, err)
	}
	defer cursor.Close(ctx)

	var orders []OrderDoc
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode KDS orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus transitions an order to a new state.
// Tiger Style: Robust ID validation and atomic status updates.
func UpdateOrderStatus(ctx context.Context, db *mongo.Database, orderID string, newStatus OrderStatus) error {
	oid, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return fmt.Errorf("invalid order ID format '%s': %w", orderID, err)
	}

	filter := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"status":     newStatus,
			"updated_at": time.Now(),
		},
	}

	_, err = db.Collection("orders").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order %s to %s: %w", orderID, newStatus, err)
	}

	return nil
}

// GetNewOrdersCount returns the count of PENDING orders.
// Tiger Style: Optimized count operation for polling.
func GetNewOrdersCount(ctx context.Context, db *mongo.Database, vendorID string) (int64, error) {
	filter := bson.M{
		"vendor_id": vendorID,
		"status":    StatusPending,
	}

	count, err := db.Collection("orders").CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count new orders for vendor %s: %w", vendorID, err)
	}

	return count, nil
}

// GetNextStatus determines the next logical step in the KDS workflow.
func GetNextStatus(current OrderStatus) OrderStatus {
	switch current {
	case StatusPending:
		return StatusConfirmed
	case StatusConfirmed:
		return StatusPreparing
	case StatusPreparing:
		return StatusReady
	case StatusReady:
		return StatusServed
	default:
		return current
	}
}
