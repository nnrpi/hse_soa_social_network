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

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"

	"social-network/statistics-service/api"
	"social-network/statistics-service/consumer"
	"social-network/statistics-service/repository"
	"social-network/statistics-service/service"
)

func main() {
	port := getEnv("PORT", "8095")
	clickhouseHost := getEnv("CLICKHOUSE_HOST", "clickhouse")
	clickhousePort := getEnv("CLICKHOUSE_PORT", "9000")
	clickhouseDB := getEnv("CLICKHOUSE_DB", "statistics")
	clickhouseUser := getEnv("CLICKHOUSE_USER", "default")
	clickhousePassword := getEnv("CLICKHOUSE_PASSWORD", "")
	kafkaBroker := getEnv("KAFKA_BROKER", "kafka:9092")
	kafkaConsumerGroup := getEnv("KAFKA_CONSUMER_GROUP", "statistics-service")

	maxRetries := 5

	db := connectClickHouse(clickhouseHost, clickhousePort, clickhouseDB, clickhouseUser, clickhousePassword, maxRetries)
	defer db.Close()

	likesRepo := repository.NewLikesRepository(db)
	commentsRepo := repository.NewCommentsRepository(db)
	viewsRepo := repository.NewViewsRepository(db)

	if err := likesRepo.Init(); err != nil {
		log.Fatalf("Failed to initialize likes table: %v", err)
	}
	if err := commentsRepo.Init(); err != nil {
		log.Fatalf("Failed to initialize comments table: %v", err)
	}
	if err := viewsRepo.Init(); err != nil {
		log.Fatalf("Failed to initialize views table: %v", err)
	}

	statsService := service.NewStatsService(likesRepo, commentsRepo, viewsRepo)

	consumerGroup := connectKafkaConsumerGroup(kafkaBroker, kafkaConsumerGroup, maxRetries)
	defer consumerGroup.Close()

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	go runConsumer(consumerCtx, consumerGroup, statsService)

	statsHandler := api.NewStatsHandler(statsService)
	router := mux.NewRouter()
	statsHandler.RegisterRoutes(router)
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Statistics service starting on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	shutdownGracefully(server, cancelConsumer)
}

func connectClickHouse(host, port, database, user, password string, maxRetries int) *sql.DB {
	var db *sql.DB
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to ClickHouse (attempt %d/%d)...", i+1, maxRetries)
		db = clickhouse.OpenDB(&clickhouse.Options{
			Addr: []string{host + ":" + port},
			Auth: clickhouse.Auth{
				Database: database,
				Username: user,
				Password: password,
			},
		})

		if err := db.Ping(); err == nil {
			log.Println("Successfully connected to ClickHouse!")
			return db
		} else {
			log.Printf("Failed to connect to ClickHouse: %v", err)
		}

		if i < maxRetries-1 {
			retryDelay := time.Duration(2<<uint(i)) * time.Second
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		} else {
			log.Fatalf("Could not connect to ClickHouse after %d attempts", maxRetries)
		}
	}
	return db
}

func connectKafkaConsumerGroup(broker, groupID string, maxRetries int) sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	var consumerGroup sarama.ConsumerGroup
	var err error
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to Kafka (attempt %d/%d)...", i+1, maxRetries)
		consumerGroup, err = sarama.NewConsumerGroup([]string{broker}, groupID, config)
		if err == nil {
			log.Println("Successfully connected to Kafka!")
			return consumerGroup
		}

		log.Printf("Failed to connect to Kafka: %v", err)
		if i < maxRetries-1 {
			retryDelay := time.Duration(2<<uint(i)) * time.Second
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		} else {
			log.Fatalf("Could not connect to Kafka after %d attempts", maxRetries)
		}
	}
	return consumerGroup
}

func runConsumer(ctx context.Context, consumerGroup sarama.ConsumerGroup, statsService *service.StatsService) {
	handler := consumer.NewHandler(statsService)

	go func() {
		for err := range consumerGroup.Errors() {
			log.Printf("Kafka consumer group error: %v", err)
		}
	}()

	for {
		if err := consumerGroup.Consume(ctx, consumer.Topics, handler); err != nil {
			log.Printf("Error from consumer group: %v", err)
		}
		if ctx.Err() != nil {
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

func shutdownGracefully(server *http.Server, cancelConsumer context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down gracefully...")
	cancelConsumer()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Statistics service stopped gracefully")
}
