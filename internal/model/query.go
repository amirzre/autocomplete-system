package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Query represents a search query with frequency tracking.
type Query struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Text      string             `json:"text" bson:"text" validate:"required,min=1,max=100"`
	Frequency int64              `json:"frequency" bson:"frequency"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
