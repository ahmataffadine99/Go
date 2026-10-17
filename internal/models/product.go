package models

import "math"

const TVARate = 0.20

func ComputeTTC(priceHT float64) float64 {
	return math.Round(priceHT*(1+TVARate)*100) / 100
}

type Product struct {
	ID          int64   `json:"id"`
	BusinessID  string  `json:"business_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	PriceTTC    float64 `json:"price_ttc"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
	CreatedAt   string  `json:"created_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

type ProductFilter struct {
	Query    string  `json:"query"`
	Category string  `json:"category"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	MinPriceTTC float64 `json:"min_price_ttc"`
	MaxPriceTTC float64 `json:"max_price_ttc"`
}
