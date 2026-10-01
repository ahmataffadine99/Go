package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/utils"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/charge"
	"github.com/stripe/stripe-go/v78/token"
)

type CartHandler struct {
	cartRepo    *repository.CartRepository
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	stripeKey   string
}

func NewCartHandler(cartRepo *repository.CartRepository, orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, stripeKey string) *CartHandler {
	return &CartHandler{
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		stripeKey:   stripeKey,
	}
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

	product, err := h.productRepo.GetByID(req.ProductID)
	if err != nil {
		http.Error(w, `{"error":"product not found"}`, http.StatusNotFound)
		return
	}

	currentQty := 0
	for _, item := range cart.Items {
		if item.ProductID == req.ProductID {
			currentQty = item.Quantity
			break
		}
	}

	if product.Stock < currentQty+req.Quantity {
		http.Error(w, `{"error":"insufficient stock"}`, http.StatusBadRequest)
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

		product, err := h.productRepo.GetByID(prodID)
		if err != nil {
			http.Error(w, `{"error":"product not found"}`, http.StatusNotFound)
			return
		}

		if product.Stock < req.Quantity {
			http.Error(w, `{"error":"insufficient stock"}`, http.StatusBadRequest)
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

	if h.stripeKey != "" {
		stripe.Key = h.stripeKey
		parts := strings.Split(req.ExpiryDate, "/")

		amountCents := int64(cart.TotalTTC * 100)

		tokenParams := &stripe.TokenParams{
			Card: &stripe.CardParams{
				Number:   stripe.String(req.CardNumber),
				ExpMonth: stripe.String(parts[0]),
				ExpYear:  stripe.String(parts[1]),
				CVC:      stripe.String(req.CVC),
			},
		}

		tok, err := token.New(tokenParams)
		if err != nil {
			http.Error(w, `{"error":"stripe token rejected: `+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		chargeParams := &stripe.ChargeParams{
			Amount:      stripe.Int64(amountCents),
			Currency:    stripe.String(string(stripe.CurrencyEUR)),
			Description: stripe.String("Commande CLI E-commerce"),
			Source:      &stripe.PaymentSourceSourceParams{Token: stripe.String(tok.ID)},
		}

		_, err = charge.New(chargeParams)
		if err != nil {
			http.Error(w, `{"error":"stripe payment failed: `+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
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
