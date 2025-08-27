package storage

import (
	"context"
	"fmt"

	"github.com/amirzre/autocomplete-system/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents the MongoDB storage implementation.
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	config     *config.Config
}

// StorageInterface defines the contract for storage operations.
type StorageInterface interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
}

// NewMongoDB creates a new MongoDB storage instance.
func NewMongoDB(config *config.Config) *MongoDB {
	return &MongoDB{
		config: config,
	}
}

// Connect establishes connection to MongoDB.
func (m *MongoDB) Connect(ctx context.Context) error {
	clientOptions := options.Client().ApplyURI(m.config.DataBase.URI)
	clientOptions.SetConnectTimeout(m.config.DataBase.ConnectTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.DataBase.Name)
	m.collection = m.database.Collection(m.config.DataBase.QueryCollection)

	if err := m.createIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// Disconnect closes the MongoDB connection.
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.client == nil {
		return nil
	}

	return m.client.Disconnect(ctx)
}

// createIndexes creates necessary indexes for optimal performance.
func (m *MongoDB) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "text", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "frequency", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: 1}},
		},
	}

	_, err := m.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}
