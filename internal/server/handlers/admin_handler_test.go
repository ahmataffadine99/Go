package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ecommerce-cli/internal/config"
	"ecommerce-cli/internal/database"
	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/server"
	"ecommerce-cli/internal/utils"
)

func doJSON(t *testing.T, method, url, token string, payload interface{}) *http.Response {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		json.NewEncoder(&body).Encode(payload)
	}
	req, _ := http.NewRequest(method, url, &body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestAdminFlow(t *testing.T) {
	db, err := database.InitDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := database.RunMigrations(db, "../../database/schema.sql"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	cfg := &config.Config{JWTSecret: "secret"}
	if err := repository.NewUserRepository(db).EnsureAdmin("admin@test.com", utils.HashPassword("admin")); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	ts := httptest.NewServer(server.NewRouter(db, cfg))
	defer ts.Close()

	resp := doJSON(t, http.MethodPost, ts.URL+"/api/auth/login", "", map[string]string{"email": "admin@test.com", "password": "admin"})
	var login struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&login)
	if resp.StatusCode != http.StatusOK || login.Token == "" {
		t.Fatalf("admin login failed: %v", resp.Status)
	}

	resp = doJSON(t, http.MethodPost, ts.URL+"/api/auth/register", "", map[string]string{"email": "client@test.com", "password": "pass"})
	var reg struct {
		UserID int64 `json:"user_id"`
	}
	json.NewDecoder(resp.Body).Decode(&reg)
	if reg.UserID == 0 {
		t.Fatalf("expected user id on register")
	}

	resp = doJSON(t, http.MethodPost, ts.URL+"/api/admin/users/confirm?id="+itoa(reg.UserID), login.Token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on confirm, got %v", resp.Status)
	}

	resp = doJSON(t, http.MethodPut, ts.URL+"/api/admin/users/update?id="+itoa(reg.UserID), login.Token, map[string]string{"email": "client2@test.com", "role": "client"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update, got %v", resp.Status)
	}

	resp = doJSON(t, http.MethodPost, ts.URL+"/api/admin/products", login.Token, map[string]interface{}{"name": "Test", "price": 10.0, "category": "Divers", "stock": 3})
	var product models.Product
	json.NewDecoder(resp.Body).Decode(&product)
	if resp.StatusCode != http.StatusCreated || product.ID == 0 || len(product.BusinessID) != 10 {
		t.Fatalf("product creation failed: %v %+v", resp.Status, product)
	}

	resp = doJSON(t, http.MethodPost, ts.URL+"/api/admin/orders/create", login.Token, map[string]interface{}{
		"user_id": reg.UserID,
		"items":   []map[string]int64{{"product_id": product.ID, "quantity": 2}},
	})
	var order models.Order
	json.NewDecoder(resp.Body).Decode(&order)
	if resp.StatusCode != http.StatusCreated || order.Status != models.OrderStatusPending || order.TotalTTC != 24 {
		t.Fatalf("order creation failed: %v %+v", resp.Status, order)
	}

	resp = doJSON(t, http.MethodPost, ts.URL+"/api/admin/orders/create", login.Token, map[string]interface{}{
		"user_id": reg.UserID,
		"items":   []map[string]int64{{"product_id": product.ID, "quantity": 2}},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected insufficient stock, got %v", resp.Status)
	}

	resp = doJSON(t, http.MethodPut, ts.URL+"/api/admin/orders/status?id="+itoa(order.ID), login.Token, map[string]string{"status": "annulé"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 when cancel reason missing, got %v", resp.Status)
	}

	resp = doJSON(t, http.MethodPut, ts.URL+"/api/admin/orders/status?id="+itoa(order.ID), login.Token, map[string]string{"status": "annulé", "cancel_reason": "rupture"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on cancel, got %v", resp.Status)
	}

	resp = doJSON(t, http.MethodDelete, ts.URL+"/api/admin/users/delete?id="+itoa(reg.UserID), login.Token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %v", resp.Status)
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
