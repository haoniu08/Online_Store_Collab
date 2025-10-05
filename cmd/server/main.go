package main

import (
	"CS6650_Online_Store/internal/handlers"
	"CS6650_Online_Store/internal/store"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize store
	productStore := store.NewProductStore()

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productStore)

	// Setup router
	router := mux.NewRouter()

	// Product endpoints
	router.HandleFunc("/products/{productId}", productHandler.GetProduct).Methods("GET")
	router.HandleFunc("/products/{productId}/details", productHandler.AddProductDetails).Methods("POST")

	// Health check endpoint (useful for load balancer)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Logging middleware
	router.Use(loggingMiddleware)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
