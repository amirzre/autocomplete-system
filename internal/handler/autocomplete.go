package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amirzre/autocomplete-system/internal/cache"
	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/model"
	"github.com/amirzre/autocomplete-system/internal/storage"
	"github.com/amirzre/autocomplete-system/internal/trie"
	"github.com/gin-gonic/gin"
)

// AutocompleteHandler handles autocomplete-related HTTP requests.
type AutocompleteHandler struct {
	trie      *trie.Trie
	storage   storage.StorageInterface
	cache     cache.CacheInterface
	config    *config.Config
	startTime time.Time
}

// NewAutocompleteHandler creates a new autocomplete handler.
func NewAutocompleteHandler(
	trie *trie.Trie,
	storage storage.StorageInterface,
	cache cache.CacheInterface,
	config *config.Config,
) *AutocompleteHandler {
	return &AutocompleteHandler{
		trie:      trie,
		storage:   storage,
		cache:     cache,
		config:    config,
		startTime: time.Now(),
	}
}

// SubmitQuery handles POST /api/v1/queries.
func (h *AutocompleteHandler) SubmitQuery(c *gin.Context) {
	var req model.SubmitQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request format",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Query cannot be empty",
			Code:    http.StatusBadRequest,
			Message: "Please provide a valid search query",
		})
		return
	}

	if len(query) > 100 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Query too long",
			Code:    http.StatusBadRequest,
			Message: "Query must be 100 characters or less",
		})
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), h.config.DataBase.QueryTimeout)
		defer cancel()

		if _, err := h.storage.UpdateQueryFrequency(ctx, query); err != nil {
			log.Printf("Error updating query frequency in storage: %v", err)
		}
	}()

	h.trie.Insert(query, 1)

	h.invalidateCacheForQuery(query)

	frequency, _ := h.trie.Search(query)

	c.JSON(http.StatusOK, model.SubmitQueryResponse{
		Message:   "Query submitted successfully",
		Query:     query,
		Frequency: frequency,
	})
}

// generateCacheKey creates a cache key for prefix and limit combination.
func (h *AutocompleteHandler) generateCacheKey(prefix string, limit int) string {
	return strings.ToLower(prefix) + ":" + strconv.Itoa(limit)
}

// invalidateCacheForQuery removes cache entries that might be affected by the new query.
func (h *AutocompleteHandler) invalidateCacheForQuery(query string) {
	query = strings.ToLower(query)

	// Invalidate cache for all prefixes of this query
	for i := 1; i <= len(query); i++ {
		prefix := query[:i]
		for limit := 1; limit <= h.config.App.MaxSuggestionsLimit; limit++ {
			cacheKey := h.generateCacheKey(prefix, limit)
			h.cache.Delete(cacheKey)
		}
	}
}
