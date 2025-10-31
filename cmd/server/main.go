package main

import (
	"CS6650_Online_Store/internal/handlers"
	"CS6650_Online_Store/internal/store"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize in-memory product store
	productStore := store.NewProductStore()

	// Initialize DB (MySQL) for shopping carts
	db, err := connectMySQLFromEnv()
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	cartRepo := store.NewMySQLCartRepository(db)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productStore)
	orderHandler := handlers.NewOrderHandler()
	cartHandler := handlers.NewShoppingCartHandler(cartRepo)

	// Setup router
	router := mux.NewRouter()

	// Order endpoints for Homework 7
	router.HandleFunc("/orders/sync", orderHandler.ProcessOrderSync).Methods("POST")
	router.HandleFunc("/orders/async", orderHandler.ProcessOrderAsync).Methods("POST")

	// Product endpoints - order matters! Specific routes before parameterized ones
	// Search endpoint for Homework 6 - searches exactly 100 products per request
	router.HandleFunc("/products/search", productHandler.SearchProducts).Methods("GET")

	router.HandleFunc("/products/{productId}", productHandler.GetProduct).Methods("GET")
	router.HandleFunc("/products/{productId}/details", productHandler.AddProductDetails).Methods("POST")

	// Shopping cart endpoints (Homework 8)
	router.HandleFunc("/shopping-carts", cartHandler.CreateCart).Methods("POST")
	router.HandleFunc("/shopping-carts/{shoppingCartId}/items", cartHandler.AddItems).Methods("POST")
	router.HandleFunc("/shopping-carts/{id}", cartHandler.GetCart).Methods("GET")

	// Health check endpoint with circuit breaker status
	router.HandleFunc("/health", productHandler.HealthCheck).Methods("GET")

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

func connectMySQLFromEnv() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "appdb"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "appuser"
	}
	pass := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4,utf8", user, pass, host, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Pooling
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
