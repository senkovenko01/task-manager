package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"task-manager/internal/config"
	"task-manager/internal/handler"
	"task-manager/internal/migrations"
	"task-manager/internal/repository"
	"task-manager/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("sqlite3", cfg.SQLitePath+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if err := migrations.Run(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Seed data if SEED_DATA environment variable is set
	if os.Getenv("SEED_DATA") == "true" {
		// Check if database already has tasks
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
		if err != nil {
			log.Printf("warning: failed to check existing tasks: %v", err)
		} else if count == 0 {
			log.Println("seeding database with sample tasks...")
			if err := migrations.SeedTasks(db); err != nil {
				log.Printf("warning: failed to seed data: %v", err)
			} else {
				log.Println("database seeded successfully")
			}
		} else {
			log.Printf("database already contains %d tasks, skipping seed", count)
		}
	}

	taskRepository := repository.NewSQLiteTaskRepository(db)
	taskService := service.NewTaskService(taskRepository)
	taskHandler := handler.NewTaskHandler(taskService)

	router := http.NewServeMux()
	taskHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		log.Println("shutting down http server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
