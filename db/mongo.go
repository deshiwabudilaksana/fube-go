package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoOnce   sync.Once
	mongoErr    error
)

// GetConnection returns a singleton MongoDB client instance.
// It initializes the connection pool once and pings the server to ensure availability.
// All subsequent calls return the same client or the initialization error.
// Tiger Style: Explicit error wrapping, context-aware initialization, and verified connectivity.
func GetConnection(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	mongoOnce.Do(func() {
		if cfg.MongoURI == "" {
			mongoErr = fmt.Errorf("MONGO_URI is not set in configuration")
			return
		}

		// Configure client options
		clientOptions := options.Client().ApplyURI(cfg.MongoURI)

		// Use a dedicated timeout context for the connection and ping phase.
		connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Establish connection to the MongoDB cluster.
		client, err := mongo.Connect(connectCtx, clientOptions)
		if err != nil {
			mongoErr = fmt.Errorf("failed to connect to MongoDB: %w", err)
			return
		}

		// Immediately verify the connection by pinging the server.
		if err := client.Ping(connectCtx, nil); err != nil {
			mongoErr = fmt.Errorf("failed to ping MongoDB: %w", err)
			// Best effort cleanup if ping fails.
			_ = client.Disconnect(ctx)
			return
		}

		mongoClient = client
	})

	if mongoErr != nil {
		return nil, mongoErr
	}

	if mongoClient == nil {
		return nil, fmt.Errorf("mongo client initialization failed silently")
	}

	return mongoClient, nil
}

// GetDatabase returns the specific MongoDB database instance configured in the app.
func GetDatabase(ctx context.Context, cfg *config.Config) (*mongo.Database, error) {
	client, err := GetConnection(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client.Database(cfg.MongoDBName), nil
}
