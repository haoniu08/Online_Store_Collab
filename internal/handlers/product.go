package handlers

import (
	"CS6650_Online_Store/internal/models"
	"CS6650_Online_Store/internal/store"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	store *store.ProductStore
}

// NewProductHandler creates a new product handler
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
