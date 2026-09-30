package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/utils"
)

type CartHandler struct {
	cartRepo  *repository.CartRepository
	orderRepo *repository.OrderRepository
}

func NewCartHandler(cartRepo *repository.CartRepository, orderRepo *repository.OrderRepository) *CartHandler {
	return &CartHandler{cartRepo: cartRepo, orderRepo: orderRepo}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(models.UserIDKey).(int64)
	cart, err := h.cartRepo.GetOrCreateActiveCart(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch cart"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(models.UserIDKey).(int64)
	cart, err := h.cartRepo.GetOrCreateActiveCart(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to get active cart"}`, http.StatusInternalServerError)
		return
	}

	var req models.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID <= 0 || req.Quantity <= 0 {
		http.Error(w, `{"error":"invalid item payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.cartRepo.AddOrUpdateItem(cart.ID, req.ProductID, req.Quantity); err != nil {
		http.Error(w, `{"error":"failed to update cart item"}`, http.StatusInternalServerError)
		return
	}

	updatedCart, _ := h.cartRepo.GetOrCreateActiveCart(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCart)
}

func (h *CartHandler) UpdateOrRemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(models.UserIDKey).(int64)
	cart, err := h.cartRepo.GetOrCreateActiveCart(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to get active cart"}`, http.StatusInternalServerError)
		return
	}

	prodIDStr := r.URL.Query().Get("product_id")
	prodID, err := strconv.ParseInt(prodIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid product_id"}`, http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		if err := h.cartRepo.RemoveItem(cart.ID, prodID); err != nil {
			http.Error(w, `{"error":"failed to delete item"}`, http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPut {
		var req models.UpdateCartItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
			return
		}
		if err := h.cartRepo.SetItemQuantity(cart.ID, prodID, req.Quantity); err != nil {
			http.Error(w, `{"error":"failed to update quantity"}`, http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	updatedCart, _ := h.cartRepo.GetOrCreateActiveCart(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCart)
}

func (h *CartHandler) Pay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(models.UserIDKey).(int64)
	cart, err := h.cartRepo.GetOrCreateActiveCart(userID)
	if err != nil || len(cart.Items) == 0 {
		http.Error(w, `{"error":"cart is empty or not found"}`, http.StatusBadRequest)
		return
	}

	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		!utils.ValidateCardNumber(req.CardNumber) ||
		!utils.ValidateCVC(req.CVC) ||
		!utils.ValidateExpiry(req.ExpiryDate) {
		http.Error(w, `{"error":"invalid card details"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.CreateFromCart(userID, cart.ID, cart.Items, cart.TotalTTC)
	if err != nil {
		http.Error(w, `{"error":"payment failed or order creation error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
