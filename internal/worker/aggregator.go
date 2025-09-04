package worker

import (
	"sync"

	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/storage"
	"github.com/amirzre/autocomplete-system/internal/trie"
)

// Aggregator handles background data aggregation tasks.
type Aggregator struct {
	trie    *trie.Trie
	storage storage.StorageInterface
	config  *config.WorkerConfig
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
	mutex   sync.Mutex
}

// NewAggregator creates a new aggregator instance.
func NewAggregator(
	trie *trie.Trie,
	storage storage.StorageInterface,
	config *config.WorkerConfig,
) *Aggregator {
	return &Aggregator{
		trie:    trie,
		storage: storage,
		config:  config,
		stopCh:  make(chan struct{}),
	}
}
