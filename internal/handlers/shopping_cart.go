package handlers

import (
	"CS6650_Online_Store/internal/store"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type ShoppingCartHandler struct {
	repo store.CartRepository
}

func NewShoppingCartHandler(repo store.CartRepository) *ShoppingCartHandler {
	return &ShoppingCartHandler{repo: repo}
}

// POST /shopping-carts
func (h *ShoppingCartHandler) CreateCart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID int64 `json:"customer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.CustomerID < 1 {
		http.Error(w, `{"error":"INVALID_INPUT","message":"invalid customer_id"}`, http.StatusBadRequest)
		return
	}
	id, err := h.repo.CreateCart(r.Context(), body.CustomerID)
	if err != nil {
		http.Error(w, `{"error":"SERVER_ERROR","message":"failed to create cart"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int64{"shopping_cart_id": id})
}

// POST /shopping-carts/{shoppingCartId}/items
func (h *ShoppingCartHandler) AddItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartIDStr := vars["shoppingCartId"]
	cartID, err := strconv.ParseInt(cartIDStr, 10, 64)
	if err != nil || cartID < 1 {
		http.Error(w, `{"error":"INVALID_INPUT","message":"invalid shoppingCartId"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProductID < 1 || body.Quantity < 1 {
		http.Error(w, `{"error":"INVALID_INPUT","message":"invalid product_id or quantity"}`, http.StatusBadRequest)
		return
	}
	if err := h.repo.AddOrUpdateItem(r.Context(), cartID, body.ProductID, body.Quantity); err != nil {
		if err == store.ErrCartNotFound {
			http.Error(w, `{"error":"NOT_FOUND","message":"shopping cart not found"}`, http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "cart not open") {
			http.Error(w, `{"error":"INVALID_INPUT","message":"invalid shopping cart state"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"SERVER_ERROR","message":"failed to add items"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /shopping-carts/{id}
func (h *ShoppingCartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		idStr = vars["shoppingCartId"]
	}
	cartID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || cartID < 1 {
		http.Error(w, `{"error":"INVALID_INPUT","message":"invalid cart id"}`, http.StatusBadRequest)
		return
	}
	cart, err := h.repo.GetCart(r.Context(), cartID)
	if err != nil {
		if err == store.ErrCartNotFound {
			http.Error(w, `{"error":"NOT_FOUND","message":"shopping cart not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"SERVER_ERROR","message":"failed to get cart"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cart)
}

// no-op
