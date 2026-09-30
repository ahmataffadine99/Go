package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-cli/internal/database"
	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/server"
)

func TestProductSearch(t *testing.T) {
	db, err := database.InitDB("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE products (
		id INTEGER PRIMARY KEY,
		business_id VARCHAR(100) UNIQUE NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DOUBLE PRECISION NOT NULL,
		category VARCHAR(100) NOT NULL,
		stock INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	INSERT INTO products (business_id, name, description, price, category, stock) 
	VALUES ('PDT-111111', 'Go Laptop', 'High performance laptop', 1200.00, 'Informatique', 5);
	INSERT INTO products (business_id, name, description, price, category, stock) 
	VALUES ('PDT-222222', 'Go Mouse', 'Wireless mouse', 25.50, 'Accessoires', 20);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to execute schema: %v", err)
	}

	router := server.NewRouter(db, "sqlite")
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Test 1: List all products
	resp, err := http.Get(ts.URL + "/api/products")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on list products, got %v", resp.Status)
	}

	var products []models.Product
	json.NewDecoder(resp.Body).Decode(&products)
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}

	// Test 2: Search by query parameter
	resp, err = http.Get(ts.URL + "/api/products?q=laptop")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on search laptop")
	}

	var searchRes []models.Product
	json.NewDecoder(resp.Body).Decode(&searchRes)
	if len(searchRes) != 1 || searchRes[0].BusinessID != "PDT-111111" {
		t.Fatalf("expected 1 product (Go Laptop), got %d", len(searchRes))
	}

	// Test 3: Search by price range
	resp, err = http.Get(ts.URL + "/api/products?min_price=10&max_price=50")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on price search")
	}

	var priceRes []models.Product
	json.NewDecoder(resp.Body).Decode(&priceRes)
	if len(priceRes) != 1 || priceRes[0].BusinessID != "PDT-222222" {
		t.Fatalf("expected 1 product (Go Mouse), got %d", len(priceRes))
	}
}
