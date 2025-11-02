package store

import (
	"CS6650_Online_Store/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCartNotFound = errors.New("shopping cart not found")
)

type CartRepository interface {
	CreateCart(ctx context.Context, customerID int64) (int64, error)
	AddOrUpdateItem(ctx context.Context, cartID int64, productID int64, quantity int) error
	GetCart(ctx context.Context, cartID int64) (*models.Cart, error)
}

type MySQLCartRepository struct {
	db *sql.DB
}

func NewMySQLCartRepository(db *sql.DB) *MySQLCartRepository {
	return &MySQLCartRepository{db: db}
}

func (r *MySQLCartRepository) CreateCart(ctx context.Context, customerID int64) (int64, error) {
	const insertCart = `INSERT INTO shopping_carts (customer_id, status) VALUES (?, 'OPEN')`
	result, err := r.db.ExecContext(ctx, insertCart, customerID)
	if err != nil {
		return 0, fmt.Errorf("insert cart: %w", err)
	}
	cartID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return cartID, nil
}

func (r *MySQLCartRepository) AddOrUpdateItem(ctx context.Context, cartID int64, productID int64, quantity int) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// rollback on panic or unhandled error
		_ = tx.Rollback()
	}()

	// Ensure cart exists and is OPEN
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM shopping_carts WHERE shopping_cart_id = ?`, cartID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCartNotFound
		}
		return fmt.Errorf("select cart: %w", err)
	}
	if status != "OPEN" {
		return fmt.Errorf("cart not open: %s", status)
	}

	// Upsert item quantity (replace quantity with provided value)
	const upsert = `INSERT INTO shopping_cart_items (shopping_cart_id, product_id, quantity)
                  VALUES (?, ?, ?)
                  ON DUPLICATE KEY UPDATE quantity = VALUES(quantity), updated_at = CURRENT_TIMESTAMP`
	if _, err := tx.ExecContext(ctx, upsert, cartID, productID, quantity); err != nil {
		return fmt.Errorf("upsert item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *MySQLCartRepository) GetCart(ctx context.Context, cartID int64) (*models.Cart, error) {
	const cartQuery = `SELECT shopping_cart_id, customer_id, status, created_at, updated_at
                     FROM shopping_carts WHERE shopping_cart_id = ?`
	var (
		id        int64
		customer  int64
		status    string
		createdAt time.Time
		updatedAt time.Time
	)
	if err := r.db.QueryRowContext(ctx, cartQuery, cartID).Scan(&id, &customer, &status, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("get cart: %w", err)
	}

	const itemsQuery = `SELECT product_id, quantity FROM shopping_cart_items WHERE shopping_cart_id = ?`
	rows, err := r.db.QueryContext(ctx, itemsQuery, cartID)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	items := make([]models.CartItem, 0, 8)
	for rows.Next() {
		var productID int64
		var quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, models.CartItem{ProductID: productID, Quantity: quantity})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return &models.Cart{
		ShoppingCartID: id,
		CustomerID:     customer,
		Status:         status,
		Items:          items,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}
