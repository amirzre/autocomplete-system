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

// GetAutocompleteSuggestions handles GET /api/v1/autocomplete.
func (h *AutocompleteHandler) GetAutocompleteSuggestions(c *gin.Context) {
	prefix := strings.TrimSpace(c.Query("q"))
	limitStr := c.Query("limit")

	if prefix == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Missing query parameter",
			Code:    http.StatusBadRequest,
			Message: "Please provide 'q' parameter with search prefix",
		})
		return
	}

	if len(prefix) > 50 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Prefix too long",
			Code:    http.StatusBadRequest,
			Message: "Search prefix must be 50 characters or less",
		})
		return
	}

	limit := h.config.App.DefaultSuggestions
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= h.config.App.MaxSuggestionsLimit {
				limit = parsedLimit
			} else if parsedLimit > h.config.App.MaxSuggestionsLimit {
				limit = h.config.App.MaxSuggestionsLimit
			}
		}
	}

	cacheKey := h.generateCacheKey(prefix, limit)
	if suggestions, found := h.cache.Get(cacheKey); found {
		c.JSON(http.StatusOK, model.AutocompleteResponse{
			Suggestions: suggestions,
			Prefix:      prefix,
			Count:       len(suggestions),
		})
		return
	}

	suggestions := h.trie.GetSuggestions(prefix, limit)

	h.cache.Set(cacheKey, suggestions)

	c.JSON(http.StatusOK, model.AutocompleteResponse{
		Suggestions: suggestions,
		Prefix:      prefix,
		Count:       len(suggestions),
	})
}

// GetStats handles GET /api/v1/stats.
func (h *AutocompleteHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	totalQueries, err := h.storage.GetTotalQueryCount(ctx)
	if err != nil {
		log.Printf("Error getting total query count: %v", err)
		totalQueries = 0
	}

	uniqueQueries, err := h.storage.GetUniqueQueryCount(ctx)
	if err != nil {
		log.Printf("Error getting unique query count: %v", err)
		uniqueQueries = h.trie.GetSize()
	}

	topQueries := h.trie.GetTopQueries(10)

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	queriesLastHour, err := h.storage.GetQueriesSince(ctx, oneHourAgo)
	if err != nil {
		log.Printf("Error getting queries from last hour: %v", err)
	}

	var lastHourCount int64
	for _, query := range queriesLastHour {
		lastHourCount += query.Frequency
	}

	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, model.StatsResponse{
		TotalQueries:    totalQueries,
		UniqueQueries:   uniqueQueries,
		TopQueries:      topQueries,
		QueriesLastHour: lastHourCount,
		SystemInfo: model.SystemInfo{
			Version:   h.config.App.Version,
			StartTime: h.startTime,
			Uptime:    uptime.String(),
		},
	})
}

// HealthCheck handles GET /api/v1/health.
func (h *AutocompleteHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.storage.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"timestamp": time.Now(),
			"error":     "Database connection failed",
		})
		return
	}

	trieSize := h.trie.GetSize()

	cacheStats := h.cache.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"timestamp":   time.Now(),
		"version":     h.config.App.Version,
		"uptime":      time.Since(h.startTime).String(),
		"trie_size":   trieSize,
		"cache_stats": cacheStats,
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
