package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/model"
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
	UpdateQueryFrequency(ctx context.Context, query string) (*model.Query, error)
	GetAllQueries(ctx context.Context) ([]model.Query, error)
	GetQueriesSince(ctx context.Context, since time.Time) ([]model.Query, error)
	GetTotalQueryCount(ctx context.Context) (int64, error)
	GetUniqueQueryCount(ctx context.Context) (int64, error)
	Ping(ctx context.Context) error
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

	if m.config.DataBase.HasAuthentication() {
		credential := options.Credential{
			AuthSource: m.config.DataBase.AuthDB,
			Username:   m.config.DataBase.Username,
			Password:   m.config.DataBase.Password,
		}
		clientOptions.SetAuth(credential)
	}

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

// UpdateQueryFrequency increments the frequency of an existing query.
func (m *MongoDB) UpdateQueryFrequency(ctx context.Context, queryText string) (*model.Query, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.DataBase.ConnectTimeout)
	defer cancel()

	filter := bson.M{"text": queryText}
	update := bson.M{
		"$inc": bson.M{"frequency": 1},
		"$set": bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"text":       queryText,
			"created_at": time.Now(),
		},
	}

	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var query model.Query
	err := m.collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&query)
	if err != nil {
		return nil, fmt.Errorf("failed to update query frequency: %w", err)
	}

	return &query, nil
}

// GetAllQueries retrieves all queries from the database.
func (m *MongoDB) GetAllQueries(ctx context.Context) ([]model.Query, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.DataBase.QueryTimeout)
	defer cancel()

	cursor, err := m.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all queries: %w", err)
	}
	defer cursor.Close(ctx)

	var queries []model.Query
	if err := cursor.All(ctx, &queries); err != nil {
		return nil, fmt.Errorf("failed to decode queries: %w", err)
	}

	return queries, nil
}

// GetQueriesSince retrieves queries created since a specific time.
func (m *MongoDB) GetQueriesSince(ctx context.Context, since time.Time) ([]model.Query, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.DataBase.QueryTimeout)
	defer cancel()

	filter := bson.M{"created_at": bson.M{"$gte": since}}
	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get queries since %v: %w", since, err)
	}
	defer cursor.Close(ctx)

	var queries []model.Query
	if err := cursor.All(ctx, &queries); err != nil {
		return nil, fmt.Errorf("failed to decode queries: %w", err)
	}

	return queries, nil
}

// GetTotalQueryCount returns the total number of query submissions.
func (m *MongoDB) GetTotalQueryCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.DataBase.QueryTimeout)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{
			Key: "$group",
			Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total", Value: bson.D{{Key: "$sum", Value: "$frequency"}}},
			},
		}},
	}

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, fmt.Errorf("failed to decode total count: %w", err)
	}

	if len(result) == 0 {
		return 0, nil
	}

	if total, ok := result[0]["total"].(int64); ok {
		return total, nil
	}

	return 0, nil
}

// GetUniqueQueryCount returns the number of unique queries.
func (m *MongoDB) GetUniqueQueryCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.DataBase.QueryTimeout)
	defer cancel()

	count, err := m.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to get unique query count: %w", err)
	}

	return count, nil
}

// Ping checks the MongoDB connection.
func (m *MongoDB) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("not connected to MongoDB")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return m.client.Ping(ctx, nil)
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
