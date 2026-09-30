package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
)

type OrderHandler struct {
	orderRepo *repository.OrderRepository
	userRepo  *repository.UserRepository
}

func NewOrderHandler(orderRepo *repository.OrderRepository, userRepo *repository.UserRepository) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, userRepo: userRepo}
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

	if err := h.orderRepo.UpdateStatus(orderID, req.Status, req.CancelReason); err != nil {
		http.Error(w, `{"error":"failed to update order status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "order status updated"})
}
