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
	_ "github.com/go-sql-driver/mysql"
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

	// Run migrations on startup (idempotent)
	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
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

// runMigrations applies database schema migrations on startup
// This is idempotent - safe to run multiple times
func runMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Execute migration statements one by one
	statements := []string{
		"DROP TABLE IF EXISTS shopping_cart_items",
		"DROP TABLE IF EXISTS shopping_carts",
		`CREATE TABLE shopping_carts (
		  shopping_cart_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		  customer_id      BIGINT UNSIGNED NOT NULL,
		  status           ENUM('OPEN','CHECKED_OUT','CANCELLED') NOT NULL DEFAULT 'OPEN',
		  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (shopping_cart_id),
		  KEY idx_shopping_carts_customer_id (customer_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE shopping_cart_items (
		  shopping_cart_id BIGINT UNSIGNED NOT NULL,
		  product_id       BIGINT UNSIGNED NOT NULL,
		  quantity         INT UNSIGNED NOT NULL,
		  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		  PRIMARY KEY (shopping_cart_id, product_id),
		  CONSTRAINT fk_items_cart
		    FOREIGN KEY (shopping_cart_id)
		    REFERENCES shopping_carts (shopping_cart_id)
		    ON DELETE CASCADE,
		  CONSTRAINT chk_item_quantity CHECK (quantity > 0),
		  KEY idx_items_product_id (product_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
