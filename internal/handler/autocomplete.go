package handler

import (
	"time"

	"github.com/amirzre/autocomplete-system/internal/cache"
	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/storage"
	"github.com/amirzre/autocomplete-system/internal/trie"
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
