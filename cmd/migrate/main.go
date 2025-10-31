package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Read DB credentials from environment variables (same as ECS tasks)
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")

	if dbHost == "" {
		log.Fatal("DB_HOST environment variable is required")
	}

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Connected to database at %s:%s", dbHost, dbPort)

	// Run migration SQL
	migrationSQL := `
-- Shopping cart core tables

-- Idempotent drops for local iteration (no-op if not exists)
DROP TABLE IF EXISTS shopping_cart_items;
DROP TABLE IF EXISTS shopping_carts;

-- Carts table
CREATE TABLE shopping_carts (
  shopping_cart_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  customer_id      BIGINT UNSIGNED NOT NULL,
  status           ENUM('OPEN','CHECKED_OUT','CANCELLED') NOT NULL DEFAULT 'OPEN',
  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (shopping_cart_id),
  KEY idx_shopping_carts_customer_id (customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Cart items table
CREATE TABLE shopping_cart_items (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

	// Execute migration
	_, err = db.Exec(migrationSQL)
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	log.Println("✅ Migration completed successfully!")
	log.Println("Tables created:")
	log.Println("  - shopping_carts")
	log.Println("  - shopping_cart_items")
}
