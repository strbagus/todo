package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo/internal/auth"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/repository"
	"todo/internal/templates"
)

func main() {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8080"
	}

	defaultDB := os.Getenv("DB_PATH")
	if defaultDB == "" {
		defaultDB = "todo.db"
	}

	defaultAuthUser := os.Getenv("AUTH_USERNAME")
	defaultAuthPass := os.Getenv("AUTH_PASSWORD")

	port := flag.String("port", defaultPort, "HTTP server port")
	dbPath := flag.String("db", defaultDB, "SQLite database file path")
	authUser := flag.String("auth-user", defaultAuthUser, "Username for basic authentication (optional)")
	authPass := flag.String("auth-pass", defaultAuthPass, "Password for basic authentication (optional)")
	flag.Parse()

	log.Printf("Starting Todo & Deadline Tracker on :%s (database: %s)...", *port, *dbPath)

	authCfg := auth.Config{
		Username: *authUser,
		Password: *authPass,
	}

	if authCfg.IsEnabled() {
		log.Printf("🔐 Authentication enabled (username: %s, persistent cookie enabled)", authCfg.Username)
	} else {
		log.Println("🔓 Authentication disabled (no credentials configured)")
	}

	// Initialize SQLite database
	db, err := database.InitDB(*dbPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// Initialize repository
	repo := repository.NewItemRepo(db)

	// Parse templates
	tmpl, err := templates.ParseTemplates()
	if err != nil {
		log.Fatalf("Template parsing failed: %v", err)
	}

	// Register routes with authentication
	handler := handlers.RegisterRoutesWithAuth(repo, tmpl, authCfg)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server ready! Open http://localhost:%s in your browser.", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
