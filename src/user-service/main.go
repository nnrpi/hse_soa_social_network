package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"social-network/user-service/api"
	"social-network/user-service/repository"
	"social-network/user-service/service"
)

func main() {
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "socialnetwork")
	serverPort := getEnv("SERVER_PORT", "8000")

	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	var db *sql.DB
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", i+1, maxRetries)
		db, err = sql.Open("postgres", dbConnectionString)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to the database!")
				break
			}
		}

		log.Printf("Failed to connect to database: %v", err)
		if i < maxRetries-1 {
			retryDelay := time.Duration(2<<uint(i)) * time.Second
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		} else {
			log.Fatalf("Could not connect to the database after %d attempts", maxRetries)
		}
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	if err := userRepo.Init(); err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}

	if err := sessionRepo.Init(); err != nil {
		log.Fatalf("Failed to initialize session repository: %v", err)
	}

	userService := service.NewUserService(userRepo, sessionRepo)
	userHandler := api.NewUserHandler(userService)
	router := mux.NewRouter()
	userHandler.RegisterRoutes(router)
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", serverPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stopCleanup := make(chan struct{})
	go periodicSessionCleanup(sessionRepo, stopCleanup)

	go func() {
		log.Printf("User service starting on port %s...", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	shutdownGracefully(server, stopCleanup, db)
}

func periodicSessionCleanup(sessionRepo *repository.SessionRepository, stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	if err := sessionRepo.CleanExpiredSessions(); err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			log.Println("Cleaning up expired sessions...")
			if err := sessionRepo.CleanExpiredSessions(); err != nil {
				log.Printf("Error cleaning up expired sessions: %v", err)
			}
		case <-stop:
			log.Println("Stopping session cleanup routine")
			return
		}
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrw := newResponseWriterWrapper(w)
		next.ServeHTTP(wrw, r)
		log.Printf(
			"%s %s %d %s %s",
			r.Method,
			r.RequestURI,
			wrw.statusCode,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rww *responseWriterWrapper) WriteHeader(statusCode int) {
	rww.statusCode = statusCode
	rww.ResponseWriter.WriteHeader(statusCode)
}

func shutdownGracefully(server *http.Server, stopCleanup chan<- struct{}, db *sql.DB) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down gracefully...")
	close(stopCleanup)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("Database closure error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
