package handlers

import (
	"CS6650_Online_Store/internal/models"
	"CS6650_Online_Store/internal/store"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	store *store.ProductStore
}

// NewProductHandler creates a new product handler with circuit breaker protection
func NewProductHandler(store *store.ProductStore) *ProductHandler {
	return &ProductHandler{store: store}
}

// GetProduct handles GET /products/{productId}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Extract productId from URL
	vars := mux.Vars(r)
	productIDStr := vars["productId"]

	// Parse productId
	productID, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID < 1 {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Invalid product ID", "Product ID must be a positive integer")
		return
	}

	// Get product from store
	product, err := h.store.GetProduct(int32(productID))
	if err != nil {
		if err == store.ErrProductNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND",
				"Product not found", "No product exists with the given ID")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Internal server error", err.Error())
		return
	}

	// Return product
	respondWithJSON(w, http.StatusOK, product)
}

// AddProductDetails handles POST /products/{productId}/details
func (h *ProductHandler) AddProductDetails(w http.ResponseWriter, r *http.Request) {
	// Extract productId from URL
	vars := mux.Vars(r)
	productIDStr := vars["productId"]

	// Parse productId
	productID, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID < 1 {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Invalid product ID", "Product ID must be a positive integer")
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Failed to read request body", err.Error())
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var product models.Product
	if err := json.Unmarshal(body, &product); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Invalid JSON format", err.Error())
		return
	}

	// Ensure the product_id in body matches URL parameter
	if product.ProductID != int32(productID) {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Product ID mismatch", "Product ID in URL must match product ID in request body")
		return
	}

	// Validate product data
	if err := product.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Invalid product data", err.Error())
		return
	}

	// Add or update product
	if err := h.store.AddOrUpdateProduct(&product); err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to save product", err.Error())
		return
	}

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}

// SearchProducts handles GET /products/search?q={query}
// This is the key endpoint for Homework 6 - searches exactly 100 products per request
// Protected by circuit breaker to prevent cascade failures under high load
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			"Missing query parameter", "Query parameter 'q' is required")
		return
	}

	// Record start time for performance measurement
	startTime := time.Now()

	searchResult, err := h.store.SearchProducts(query, 100, 20)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Search failed", err.Error())
		return
	}

	// Consider very slow responses as degraded
	searchDuration := time.Since(startTime)
	if searchDuration > 2*time.Second {
		respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"Search service temporarily unavailable", "Search operation exceeded acceptable latency")
		return
	}

	searchResult.SearchTime = searchDuration.String()
	respondWithJSON(w, http.StatusOK, searchResult)
}

// HealthCheck handles GET /health - returns system health and circuit breaker status
func (h *ProductHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthResponse := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	respondWithJSON(w, http.StatusOK, healthResponse)
}

// Helper functions for responses

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, statusCode int, errorCode, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := models.NewError(errorCode, message, details)
	json.NewEncoder(w).Encode(errorResponse)
}
