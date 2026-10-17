package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-cli/internal/config"
	"ecommerce-cli/internal/database"
	"ecommerce-cli/internal/server"
)

func TestAuthFlow(t *testing.T) {
	db, err := database.InitDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'client',
		is_confirmed BOOLEAN NOT NULL DEFAULT 0,
		confirmation_code TEXT,
		reset_code TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	router := server.NewRouter(db, &config.Config{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	// 1. Inscription
	regPayload := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(regPayload)

	resp, err := http.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created on register, got status %v", resp.Status)
	}

	var regRes map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&regRes)
	confirmCode, ok := regRes["confirmation_code"].(string)
	if !ok || confirmCode == "" {
		t.Fatalf("expected valid confirmation_code in response")
	}

	// 2. Confirmation
	confPayload := map[string]string{
		"email": "test@example.com",
		"code":  confirmCode,
	}
	body, _ = json.Marshal(confPayload)

	resp, err = http.Post(ts.URL+"/api/auth/confirm", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on confirm, got status %v", resp.Status)
	}

	// 3. Connexion
	loginPayload := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayload)

	resp, err = http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on login, got status %v", resp.Status)
	}

	var loginRes map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginRes)
	token, ok := loginRes["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in login response")
	}

	// 4. Réinitialisation mot de passe
	body, _ = json.Marshal(map[string]string{"email": "test@example.com"})
	resp, err = http.Post(ts.URL+"/api/auth/forgot-password", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on forgot password, got status %v", resp.Status)
	}

	var resetCode string
	if err := db.QueryRow("SELECT reset_code FROM users WHERE email = ?", "test@example.com").Scan(&resetCode); err != nil || resetCode == "" {
		t.Fatalf("expected reset code to be stored")
	}

	resetPayload := map[string]string{
		"email":        "test@example.com",
		"code":         "WRONG1",
		"new_password": "newpassword456",
	}
	body, _ = json.Marshal(resetPayload)
	resp, err = http.Post(ts.URL+"/api/auth/reset-password", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 on reset with wrong code, got status %v", resp.Status)
	}

	resetPayload["code"] = resetCode
	body, _ = json.Marshal(resetPayload)

	resp, err = http.Post(ts.URL+"/api/auth/reset-password", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on reset password, got status %v", resp.Status)
	}

	// 5. Connexion avec le nouveau mot de passe
	loginPayload["password"] = "newpassword456"
	body, _ = json.Marshal(loginPayload)

	resp, err = http.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on login with new password, got status %v", resp.Status)
	}
}

func TestForgedTokenRejected(t *testing.T) {
	db, err := database.InitDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	router := server.NewRouter(db, &config.Config{JWTSecret: "secret"})
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer token_1_admin_0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with forged token, got %v", resp.Status)
	}
}
