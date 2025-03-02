package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	port := getEnv("PORT", "8080")
	userServiceURL := getEnv("USER_SERVICE_URL", "http://user-service:8000")

	router := mux.NewRouter()

	targetURL, err := url.Parse(userServiceURL)
	if err != nil {
		log.Fatalf("Error parsing user service URL: %v", err)
	}

	userServiceProxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := userServiceProxy.Director
	userServiceProxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Proxy", "Social Network Proxy")
		if cookies := req.Cookies(); len(cookies) > 0 {
			log.Printf("Proxying cookies: %v", cookies)
		}
	}
	userServiceProxy.ModifyResponse = func(resp *http.Response) error {
		cookies := resp.Cookies()
		if len(cookies) > 0 {
			log.Printf("Response setting cookies: %v", cookies)
		}
		return nil
	}

	router.PathPrefix("/auth/").Handler(userServiceProxy)
	router.PathPrefix("/users/").Handler(userServiceProxy)

	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Proxy service starting on %s...", addr)
	log.Printf("Proxying requests to user service at %s", userServiceURL)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"proxy service is running"}`))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		wrapper := newResponseWriterWrapper(w)
		next.ServeHTTP(wrapper, r)
		log.Printf("Response: %d %s %s [%v]",
			wrapper.statusCode,
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

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

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
