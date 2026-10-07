package models

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "en attente"
	OrderStatusPaid      OrderStatus = "payé"
	OrderStatusShipping  OrderStatus = "en cours de livraison"
	OrderStatusDelivered OrderStatus = "livré"
	OrderStatusCanceled  OrderStatus = "annulé"
)

type OrderItem struct {
	ID        int64    `json:"id"`
	OrderID   int64    `json:"order_id"`
	ProductID int64    `json:"product_id"`
	Quantity  int      `json:"quantity"`
	UnitPrice float64  `json:"unit_price"`
	Product   *Product `json:"product,omitempty"`
}

type Order struct {
	ID           int64       `json:"id"`
	BusinessID   string      `json:"business_id"`
	UserID       int64       `json:"user_id"`
	CartID       int64       `json:"cart_id"`
	TotalTTC     float64     `json:"total_ttc"`
	Status       OrderStatus `json:"status"`
	CancelReason string      `json:"cancel_reason,omitempty"`
	Items        []OrderItem `json:"items,omitempty"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
}

type UpdateOrderStatusRequest struct {
	Status       OrderStatus `json:"status"`
	CancelReason string      `json:"cancel_reason,omitempty"`
}

type AdminCreateOrderRequest struct {
	UserID     int64              `json:"user_id"`
	Items      []AddToCartRequest `json:"items"`
	CardNumber string             `json:"card_number"`
	ExpiryDate string             `json:"expiry_date"`
	CVC        string             `json:"cvc"`
}
