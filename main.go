package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"taskapi/config"
	"taskapi/database"
	"taskapi/handlers"
	"taskapi/middleware"
	"taskapi/services"
	"taskapi/worker"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v\n", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize services (use package-level repository functions)
	userService := services.NewUserService(db, cfg)
	taskService := services.NewTaskService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Start background worker
	taskWorker := worker.NewTaskWorker(db, cfg)
	taskWorker.Start()

	// Setup routes
	router := mux.NewRouter()

	// Auth routes (no authentication required)
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Protected task routes
	protectedRouter := router.PathPrefix("/api/tasks").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(cfg))

	protectedRouter.HandleFunc("", taskHandler.CreateTask).Methods("POST")
	protectedRouter.HandleFunc("", taskHandler.GetTasks).Methods("GET")
	protectedRouter.HandleFunc("/{id}", taskHandler.GetTask).Methods("GET")
	protectedRouter.HandleFunc("/{id}", taskHandler.UpdateTask).Methods("PUT")
	protectedRouter.HandleFunc("/{id}", taskHandler.DeleteTask).Methods("DELETE")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nShutting down server...")
		taskWorker.Stop()
		os.Exit(0)
	}()

	// Start server
	log.Printf("Server starting on port %s\n", cfg.ServerPort)
	log.Printf("Auto-complete delay: %d minutes\n", cfg.AutoCompleteMinutes)

	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
