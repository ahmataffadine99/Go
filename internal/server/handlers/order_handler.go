package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
)

type OrderHandler struct {
	orderRepo   *repository.OrderRepository
	userRepo    *repository.UserRepository
	productRepo *repository.ProductRepository
}

func NewOrderHandler(orderRepo *repository.OrderRepository, userRepo *repository.UserRepository, productRepo *repository.ProductRepository) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, userRepo: userRepo, productRepo: productRepo}
}

func (h *OrderHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(models.UserIDKey).(int64)
	orders, err := h.orderRepo.GetUserOrders(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch orders"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) AdminListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	orders, err := h.orderRepo.ListAllOrders()
	if err != nil {
		http.Error(w, `{"error":"failed to list orders"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	orderIDStr := r.URL.Query().Get("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid order id"}`, http.StatusBadRequest)
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if !req.Status.IsValid() {
		http.Error(w, `{"error":"invalid status"}`, http.StatusBadRequest)
		return
	}

	req.CancelReason = strings.TrimSpace(req.CancelReason)
	if req.Status == models.OrderStatusCanceled && req.CancelReason == "" {
		http.Error(w, `{"error":"a cancel reason is required"}`, http.StatusBadRequest)
		return
	}
	if req.Status != models.OrderStatusCanceled {
		req.CancelReason = ""
	}

	if err := h.orderRepo.UpdateStatus(orderID, req.Status, req.CancelReason); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"failed to update order status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "order status updated"})
}

func (h *OrderHandler) AdminCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.AdminCreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 || len(req.Items) == 0 {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if _, err := h.userRepo.GetByID(req.UserID); err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	quantities := make(map[int64]int)
	var order []int64
	for _, item := range req.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			http.Error(w, `{"error":"invalid item"}`, http.StatusBadRequest)
			return
		}
		if _, seen := quantities[item.ProductID]; !seen {
			order = append(order, item.ProductID)
		}
		quantities[item.ProductID] += item.Quantity
	}

	var items []models.OrderItem
	for _, productID := range order {
		product, err := h.productRepo.GetByID(productID)
		if err != nil {
			http.Error(w, `{"error":"product not found"}`, http.StatusNotFound)
			return
		}
		items = append(items, models.OrderItem{
			ProductID: product.ID,
			Quantity:  quantities[productID],
			UnitPrice: product.PriceTTC,
			Product:   product,
		})
	}

	created, err := h.orderRepo.CreateForUser(req.UserID, items)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientStock) {
			http.Error(w, `{"error":"insufficient stock"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"failed to create order"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
