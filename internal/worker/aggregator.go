package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/model"
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

// Start begins the aggregation worker
func (a *Aggregator) Start(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.running {
		return nil
	}

	if !a.config.Enabled {
		log.Println("worker aggregation is disabled")
		return nil
	}

	log.Printf("Starting aggregation worker with interval: %v", a.config.AggregationInterval)

	a.running = true
	a.wg.Add(1)

	go a.runAggregation(ctx)

	return nil
}

// runAggregation runs the main aggregation loop
func (a *Aggregator) runAggregation(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.AggregationInterval)
	defer ticker.Stop()

	if err := a.performAggregation(ctx); err != nil {
		log.Printf("Error during initial aggregation: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := a.performAggregation(ctx); err != nil {
				log.Printf("Error during aggregation: %v", err)
			}
		case <-a.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// performAggregation syncs data from storage to trie.
func (a *Aggregator) performAggregation(ctx context.Context) error {
	log.Println("Starting data aggregation...")
	start := time.Now()

	queries, err := a.storage.GetAllQueries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get queries form storage: %w", err)
	}

	if len(queries) == 0 {
		log.Println("No queries found in storage")
		return nil
	}

	a.trie.Clear()

	batchSize := a.config.BatchSize
	processed := 0

	for i := 0; i < len(queries); i += batchSize {
		end := i + batchSize
		if end > len(queries) {
			end = len(queries)
		}

		batch := queries[i:end]
		if err := a.processBatch(batch); err != nil {
			log.Printf("Error processing batch %d-%d: %v", i, end, err)
			continue
		}

		processed += len(batch)
	}

	duration := time.Since(start)
	log.Printf("Aggregation completed: processed %d queries in %v", processed, duration)

	return nil
}

// processBatch processes a batch of queries.
func (a *Aggregator) processBatch(queries []model.Query) error {
	for _, query := range queries {
		if query.Text != "" && query.Frequency > 0 {
			a.trie.Insert(query.Text, query.Frequency)
		}
	}
	return nil
}
