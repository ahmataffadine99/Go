package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/utils"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	confirmCode := utils.GenerateConfirmationCode()
	user := &models.User{
		Email:            req.Email,
		PasswordHash:     utils.HashPassword(req.Password),
		Role:             "client",
		IsConfirmed:      false,
		ConfirmationCode: confirmCode,
	}

	if err := h.userRepo.Create(user); err != nil {
		http.Error(w, `{"error":"user already exists or database error"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "user registered successfully, please confirm account",
		"confirmation_code": confirmCode,
		"user_id":           user.ID,
	})
}

func (h *AuthHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Code == "" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if err := h.userRepo.ConfirmUser(req.Email, req.Code); err != nil {
		log.Printf("Confirmation error: %v", err)
		http.Error(w, `{"error":"invalid confirmation code or email"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "account confirmed successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || !utils.VerifyPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if !user.IsConfirmed {
		http.Error(w, `{"error":"account not confirmed"}`, http.StatusForbidden)
		return
	}

	token := utils.GenerateSimpleToken(user.ID, user.Role)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	newHash := utils.HashPassword(req.NewPassword)
	if err := h.userRepo.UpdatePassword(req.Email, newHash); err != nil {
		http.Error(w, `{"error":"email not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "password reset successfully"})
}
