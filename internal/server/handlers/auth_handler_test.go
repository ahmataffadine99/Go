package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	router := server.NewRouter(db)
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
	resetPayload := map[string]string{
		"email":        "test@example.com",
		"new_password": "newpassword456",
	}
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
