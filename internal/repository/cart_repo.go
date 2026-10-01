package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"ecommerce-cli/internal/database"
	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/utils"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetOrCreateActiveCart(userID int64) (*models.Cart, error) {
	query := `SELECT id, business_id, user_id, status, created_at, updated_at 
	          FROM carts WHERE user_id = ? AND status = 'active'`
	row := r.db.QueryRow(database.RebindQuery(query), userID)

	var c models.Cart
	err := row.Scan(&c.ID, &c.BusinessID, &c.UserID, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		bID := utils.GenerateCartID()
		insertQ := `INSERT INTO carts (business_id, user_id, status) VALUES (?, ?, 'active')`
		res, err := r.db.Exec(database.RebindQuery(insertQ), bID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to create cart: %w", err)
		}
		id, _ := res.LastInsertId()
		c = models.Cart{
			ID:         id,
			BusinessID: bID,
			UserID:     userID,
			Status:     "active",
		}
	} else if err != nil {
		return nil, err
	}

	items, total, err := r.GetCartItems(c.ID)
	if err != nil {
		return nil, err
	}
	c.Items = items
	c.TotalTTC = total

	return &c, nil
}

func (r *CartRepository) AddOrUpdateItem(cartID, productID int64, quantity int) error {
	var currentQty int
	err := r.db.QueryRow(database.RebindQuery("SELECT quantity FROM cart_items WHERE cart_id = ? AND product_id = ?"), cartID, productID).Scan(&currentQty)

	if err == sql.ErrNoRows {
		_, err = r.db.Exec(database.RebindQuery("INSERT INTO cart_items (cart_id, product_id, quantity) VALUES (?, ?, ?)"), cartID, productID, quantity)
		return err
	} else if err != nil {
		return err
	}

	newQty := currentQty + quantity
	if newQty <= 0 {
		_, err = r.db.Exec(database.RebindQuery("DELETE FROM cart_items WHERE cart_id = ? AND product_id = ?"), cartID, productID)
		return err
	}

	_, err = r.db.Exec(database.RebindQuery("UPDATE cart_items SET quantity = ? WHERE cart_id = ? AND product_id = ?"), newQty, cartID, productID)
	return err
}

func (r *CartRepository) SetItemQuantity(cartID, productID int64, quantity int) error {
	if quantity <= 0 {
		_, err := r.db.Exec(database.RebindQuery("DELETE FROM cart_items WHERE cart_id = ? AND product_id = ?"), cartID, productID)
		return err
	}
	_, err := r.db.Exec(database.RebindQuery("UPDATE cart_items SET quantity = ? WHERE cart_id = ? AND product_id = ?"), quantity, cartID, productID)
	return err
}

func (r *CartRepository) RemoveItem(cartID, productID int64) error {
	_, err := r.db.Exec(database.RebindQuery("DELETE FROM cart_items WHERE cart_id = ? AND product_id = ?"), cartID, productID)
	return err
}

func (r *CartRepository) GetCartItems(cartID int64) ([]models.CartItem, float64, error) {
	query := `SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity, 
	                 p.id, p.business_id, p.name, p.description, p.price, p.category, p.stock
	          FROM cart_items ci
	          JOIN products p ON ci.product_id = p.id
	          WHERE ci.cart_id = ?`

	rows, err := r.db.Query(database.RebindQuery(query), cartID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.CartItem
	var total float64

	for rows.Next() {
		var ci models.CartItem
		var p models.Product
		if err := rows.Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.Quantity,
			&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Stock); err != nil {
			return nil, 0, err
		}
		ci.Product = &p
		items = append(items, ci)
		total += p.Price * float64(ci.Quantity)
	}

	return items, total, nil
}

func (r *CartRepository) CloseCart(cartID int64) error {
	_, err := r.db.Exec(database.RebindQuery("UPDATE carts SET status = 'checkout' WHERE id = ?"), cartID)
	return err
}

var ErrCartEmpty = errors.New("cart is empty")
