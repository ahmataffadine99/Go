package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ecommerce-cli/internal/models"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(p *models.Product) error {
	query := `INSERT INTO products (business_id, name, description, price, category, stock) 
	          VALUES (?, ?, ?, ?, ?, ?)`
	res, err := r.db.Exec(query, p.BusinessID, p.Name, p.Description, p.Price, p.Category, p.Stock)
	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		p.ID = id
	}
	return nil
}

func (r *ProductRepository) GetByID(id int64) (*models.Product, error) {
	query := `SELECT id, business_id, name, description, price, category, stock, created_at 
	          FROM products WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var p models.Product
	err := row.Scan(&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Stock, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Search(filter models.ProductFilter) ([]models.Product, error) {
	query := `SELECT id, business_id, name, description, price, category, stock, created_at FROM products WHERE 1=1`
	var args []interface{}

	if filter.Query != "" {
		query += ` AND (LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(business_id) LIKE ?)`
		term := "%" + strings.ToLower(filter.Query) + "%"
		args = append(args, term, term, term)
	}

	if filter.Category != "" {
		query += ` AND LOWER(category) = ?`
		args = append(args, strings.ToLower(filter.Category))
	}

	if filter.MinPrice > 0 {
		query += ` AND price >= ?`
		args = append(args, filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		query += ` AND price <= ?`
		args = append(args, filter.MaxPrice)
	}

	query += ` ORDER BY id DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Stock, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
