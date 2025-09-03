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

// Suggestion represents a single autocomplete suggestion.
type Suggestion struct {
	Text      string `json:"text"`
	Frequency int64  `json:"frequency"`
}

// SubmitQueryRequest represents the request to submit a new query.
type SubmitQueryRequest struct {
	Query string `json:"query" validate:"required,min=1,max=100"`
}

// SubmitQueryResponse represents the response after submitting a query.
type SubmitQueryResponse struct {
	Message   string `json:"message"`
	Query     string `json:"query"`
	Frequency int64  `json:"frequency"`
}

// AutocompleteRequest represents the request for autocomplete suggestions.
type AutocompleteRequest struct {
	Prefix string `json:"prefix" form:"q" validate:"required,min=1,max=50"`
	Limit  int    `json:"limit" form:"limit" validate:"min=1,max=20"`
}

// AutocompleteResponse represents the autocomplete API response.
type AutocompleteResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Prefix      string       `json:"prefix"`
	Count       int          `json:"count"`
}

// StatsResponse represents system statistics.
type StatsResponse struct {
	TotalQueries    int64        `json:"total_queries"`
	UniqueQueries   int64        `json:"unique_queries"`
	TopQueries      []Suggestion `json:"top_queries"`
	QueriesLastHour int64        `json:"queries_last_hour"`
	SystemInfo      SystemInfo   `json:"system_info"`
}

// SystemInfo represents system information.
type SystemInfo struct {
	Version   string    `json:"version"`
	StartTime time.Time `json:"start_time"`
	Uptime    string    `json:"uptime"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
