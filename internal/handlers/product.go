package handlers

import (
	"CS6650_Online_Store/internal/circuitbreaker"
	"CS6650_Online_Store/internal/models"
	"CS6650_Online_Store/internal/store"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	store         *store.ProductStore
	searchBreaker *circuitbreaker.CircuitBreaker
}

// NewProductHandler creates a new product handler with circuit breaker protection
func NewProductHandler(store *store.ProductStore) *ProductHandler {
	// Configure circuit breaker for search operations
	searchBreakerConfig := circuitbreaker.Config{
		FailureThreshold: 5,                // Open after 5 failures
		RecoveryTimeout:  30 * time.Second, // Wait 30 seconds before trying again
		SuccessThreshold: 3,                // Need 3 successes to close circuit
	}

	return &ProductHandler{
		store:         store,
		searchBreaker: circuitbreaker.NewCircuitBreaker(searchBreakerConfig),
	}
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
	var searchResult *models.SearchResponse

	// Execute search with circuit breaker protection
	err := h.searchBreaker.Execute(func() error {
		var searchErr error
		searchResult, searchErr = h.store.SearchProducts(query, 100, 20)

		// Consider slow responses (>2 seconds) as failures to trigger circuit breaker
		if time.Since(startTime) > 2*time.Second {
			return errors.New("search operation too slow - response time exceeded 2 seconds")
		}

		return searchErr
	})

	if err != nil {
		// Check if error is from circuit breaker
		if err.Error() == "circuit breaker is open - service temporarily unavailable" {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"Search service temporarily unavailable", "Circuit breaker is open due to high failure rate. Please try again later.")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Search failed", err.Error())
		return
	}

	// Add search time to response
	searchDuration := time.Since(startTime)
	searchResult.SearchTime = searchDuration.String()

	// Return search results
	respondWithJSON(w, http.StatusOK, searchResult)
}

// HealthCheck handles GET /health - returns system health and circuit breaker status
func (h *ProductHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	metrics := h.searchBreaker.GetMetrics()

	healthResponse := map[string]interface{}{
		"status":          "healthy",
		"timestamp":       time.Now().Format(time.RFC3339),
		"circuit_breaker": metrics,
	}

	// Return 503 if circuit breaker is open
	if h.searchBreaker.GetState() == circuitbreaker.StateOpen {
		healthResponse["status"] = "degraded"
		respondWithJSON(w, http.StatusServiceUnavailable, healthResponse)
		return
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
