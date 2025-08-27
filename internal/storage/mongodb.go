package storage

import (
	"github.com/amirzre/autocomplete-system/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDB represents the MongoDB storage implementation.
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	config     *config.Config
}

// StorageInterface defines the contract for storage operations.
type StorageInterface interface{}

// NewMongoDB creates a new MongoDB storage instance.
func NewMongoDB(config *config.Config) *MongoDB {
	return &MongoDB{
		config: config,
	}
}
