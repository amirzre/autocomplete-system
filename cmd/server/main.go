package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amirzre/autocomplete-system/internal/config"
	"github.com/gin-gonic/gin"
)

const (
	appName    = "Autocomplete System"
	appVersion = "1.0.0"
)

func main() {
	// Load configuration
	cfg := config.Load()

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
	router *gin.Engine
	cfg    *config.Config
}

// initializeApp sets up all application components.
func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{
		cfg: cfg,
	}

	app.setupRouter()
	log.Println("Routes configured")

	return app, nil
}

// setupRouter configures all HTTP routes and middleware.
func (app *App) setupRouter() {
	router := gin.New()

	app.router = router
}

// cleanup performs graceful cleanup of resources.
func (app *App) cleanup(ctx context.Context) {
	log.Println("Cleaning up resources...")
	log.Println("Cleanup completed")
}
