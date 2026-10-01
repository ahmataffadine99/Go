package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"ecommerce-cli/internal/database"
	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/utils"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateFromCart(userID, cartID int64, items []models.CartItem, totalTTC float64) (*models.Order, error) {
	if len(items) == 0 {
		return nil, ErrCartEmpty
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bID := utils.GenerateOrderID()
	orderQuery := `INSERT INTO orders (business_id, user_id, cart_id, total_ttc, status) VALUES (?, ?, ?, ?, ?)`
	res, err := tx.Exec(orderQuery, bID, userID, cartID, totalTTC, string(models.OrderStatusPaid))
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	orderID, _ := res.LastInsertId()

	for _, item := range items {
		itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)`
		_, err := tx.Exec(itemQuery, orderID, item.ProductID, item.Quantity, item.Product.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	_, err = tx.Exec(database.RebindQuery("UPDATE carts SET status = 'completed' WHERE id = ?"), cartID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Order{
		ID:         orderID,
		BusinessID: bID,
		UserID:     userID,
		CartID:     cartID,
		TotalTTC:   totalTTC,
		Status:     models.OrderStatusPaid,
	}, nil
}

func (r *OrderRepository) GetUserOrders(userID int64) ([]models.Order, error) {
	query := `SELECT id, business_id, user_id, cart_id, total_ttc, status, cancel_reason, created_at, updated_at 
	          FROM orders WHERE user_id = ? ORDER BY id DESC`
	rows, err := r.db.Query(database.RebindQuery(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var reason sql.NullString
		if err := rows.Scan(&o.ID, &o.BusinessID, &o.UserID, &o.CartID, &o.TotalTTC, &o.Status, &reason, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		if reason.Valid {
			o.CancelReason = reason.String
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) ListAllOrders() ([]models.Order, error) {
	query := `SELECT id, business_id, user_id, cart_id, total_ttc, status, cancel_reason, created_at, updated_at 
	          FROM orders ORDER BY id DESC`
	rows, err := r.db.Query(database.RebindQuery(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var reason sql.NullString
		if err := rows.Scan(&o.ID, &o.BusinessID, &o.UserID, &o.CartID, &o.TotalTTC, &o.Status, &reason, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		if reason.Valid {
			o.CancelReason = reason.String
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(orderID int64, status models.OrderStatus, reason string) error {
	query := `UPDATE orders SET status = ?, cancel_reason = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	res, err := r.db.Exec(database.RebindQuery(query), string(status), reason, orderID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrOrderNotFound
	}
	return nil
}
