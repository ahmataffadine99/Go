package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/utils"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
}

func normalizeRole(role string) (string, bool) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		role = "client"
	}
	return role, role == "client" || role == "admin"
}

func (h *UserHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	users, err := h.userRepo.ListAll()
	if err != nil {
		http.Error(w, `{"error":"failed to fetch users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.AdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	role, ok := normalizeRole(req.Role)
	if !ok {
		http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
		return
	}

	user := &models.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: utils.HashPassword(req.Password),
		Role:         role,
		IsConfirmed:  true,
	}

	if err := h.userRepo.Create(user); err != nil {
		http.Error(w, `{"error":"user already exists or database error"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := parseID(r)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var req models.AdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	role, ok := normalizeRole(req.Role)
	if !ok {
		http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Update(id, strings.ToLower(strings.TrimSpace(req.Email)), role); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"failed to update user"}`, http.StatusBadRequest)
		return
	}

	if req.Password != "" {
		if err := h.userRepo.UpdatePasswordByID(id, utils.HashPassword(req.Password)); err != nil {
			http.Error(w, `{"error":"failed to update password"}`, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user updated"})
}

func (h *UserHandler) ConfirmUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := parseID(r)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := h.userRepo.ConfirmByID(id); err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user confirmed"})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := parseID(r)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if currentID, _ := r.Context().Value(models.UserIDKey).(int64); currentID == id {
		http.Error(w, `{"error":"cannot delete your own account"}`, http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		http.Error(w, `{"error":"failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user deleted"})
}
