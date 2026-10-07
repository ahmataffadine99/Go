package models

type CartItem struct {
	ID        int64    `json:"id"`
	CartID    int64    `json:"cart_id"`
	ProductID int64    `json:"product_id"`
	Quantity  int      `json:"quantity"`
	Product   *Product `json:"product,omitempty"`
}

type Cart struct {
	ID         int64      `json:"id"`
	BusinessID string     `json:"business_id"`
	UserID     int64      `json:"user_id"`
	Status     string     `json:"status"`
	Items      []CartItem `json:"items"`
	TotalTTC   float64    `json:"total_ttc"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
}

type AddToCartRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

type PaymentRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVC        string `json:"cvc"`
}
