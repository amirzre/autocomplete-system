package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amirzre/autocomplete-system/internal/cache"
	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/amirzre/autocomplete-system/internal/handler"
	"github.com/amirzre/autocomplete-system/internal/storage"
	"github.com/amirzre/autocomplete-system/internal/trie"
	"github.com/amirzre/autocomplete-system/internal/worker"
	"github.com/gin-gonic/gin"
)

const (
	appName    = "Autocomplete System"
	appVersion = "1.0.0"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	app, err := initializeApp(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to initialize application: %w\n", err)
	}
	defer app.cleanup(ctx)

	// Start HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      app.router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting %s v%s on %s", appName, appVersion, cfg.Server.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// App holds all application components.
type App struct {
	router  *gin.Engine
	config  *config.Config
	storage storage.StorageInterface
	cache   cache.CacheInterface
	trie    *trie.Trie
	handler *handler.AutocompleteHandler
	worker  *worker.Aggregator
}

// initializeApp sets up all application components.
func initializeApp(ctx context.Context, config *config.Config) (*App, error) {
	app := &App{
		config: config,
	}

	// Initialize Trie
	app.trie = trie.New()
	log.Println("Trie initialized")

	// Initialize Storage
	mongoStorage := storage.NewMongoDB(config)
	if err := mongoStorage.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	app.storage = mongoStorage
	log.Println("MongoDB connected")

	// Initialize Redis Cache
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis cache: %w", err)
	}
	app.cache = redisCache
	log.Println("Redis cache connected")

	// Initialize Worker
	app.worker = worker.NewAggregator(app.trie, app.storage, &config.Worker)
	if err := app.worker.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start worker: %w", err)
	}
	log.Println("Background worker started")

	// Initialize Handlers
	app.handler = handler.NewAutocompleteHandler(app.trie, app.storage, app.cache, config)
	log.Println("Handler initialized")

	// Setup router
	app.setupRouter()
	log.Println("Routes configured")

	return app, nil
}

// setupRouter configures all HTTP routes and middleware.
func (app *App) setupRouter() {
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// API v1 routes
	v1 := router.Group("/api/v1")

	v1.GET("/health", app.handler.HealthCheck)

	v1.POST("/queries", app.handler.SubmitQuery)
	v1.GET("/autocomplete", app.handler.GetAutocompleteSuggestions)
	v1.GET("/stats", app.handler.GetStats)

	cache := v1.Group("/cache")
	cache.GET("/stats", app.handler.GetCacheStats)
	cache.DELETE("/", app.handler.ClearCache)

	app.router = router
}

// cleanup performs graceful cleanup of resources.
func (app *App) cleanup(ctx context.Context) {
	log.Println("Cleaning up resources...")

	// Stop worker
	if app.worker != nil {
		if err := app.worker.Stop(); err != nil {
			log.Printf("Error stopping worker: %v", err)
		}
	}

	// Disconnect from storage
	if app.storage != nil {
		if err := app.storage.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from storage: %v", err)
		}
	}

	log.Println("Cleanup completed")
}
