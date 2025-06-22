package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/gorilla/mux"
	"github.com/upamune/claude-code-pull-worker/internal/config"
	"github.com/upamune/claude-code-pull-worker/internal/database"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/handlers"
	"github.com/upamune/claude-code-pull-worker/internal/worker"
)

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "systemd-install":
		err := cli.SystemdInstall.Run()
		ctx.FatalIfErrorf(err)
		return
	default:
		runServer(cli.Server)
	}
}

func runServer(serverCmd Server) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	database, err := database.New("claude-code-pull-worker.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create queries instance
	queries := db.New(database)

	// Initialize handlers
	adminHandler, err := handlers.NewAdminHandler(queries)
	if err != nil {
		log.Fatalf("Failed to initialize admin handler: %v", err)
	}

	webhookHandler := handlers.NewWebhookExecutionHandler(queries)

	// Create and start queue worker
	queueWorker := worker.NewQueueWorker(queries)
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	go queueWorker.Start(workerCtx)
	log.Println("Queue worker started")

	// Setup routes
	r := mux.NewRouter()
	
	// Register admin routes
	adminHandler.RegisterRoutes(r)
	
	// Register webhook execution routes
	r.HandleFunc("/webhooks/{uuid}", webhookHandler.HandleWebhookExecution).Methods("POST")
	
	// Legacy endpoint (for backward compatibility)
	r.HandleFunc("/webhook", handleLegacyWebhook).Methods("POST")
	r.HandleFunc("/health", handleHealth).Methods("GET")

	// Serve static files if needed
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Admin interface available at http://localhost:%s/", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Stop worker
	cancelWorker()
	queueWorker.Stop()

	// Shutdown HTTP server
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func handleLegacyWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusGone)
	w.Write([]byte(`{"error": "This endpoint is deprecated. Please use /webhooks/{uuid} instead."}`))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "healthy"}`))
}