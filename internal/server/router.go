package server

import (
	"database/sql"
	"net/http"

	"ecommerce-cli/internal/config"
	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/server/handlers"
)

func NewRouter(db *sql.DB, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	auth := AuthMiddleware(cfg.JWTSecret)
	admin := AdminOnlyMiddleware(cfg.JWTSecret)

	authH := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)
	prodH := handlers.NewProductHandler(productRepo)
	cartH := handlers.NewCartHandler(cartRepo, orderRepo, productRepo)
	orderH := handlers.NewOrderHandler(orderRepo, userRepo, productRepo)
	userH := handlers.NewUserHandler(userRepo)

	// Public Auth Endpoints
	mux.HandleFunc("/api/auth/register", authH.Register)
	mux.HandleFunc("/api/auth/confirm", authH.Confirm)
	mux.HandleFunc("/api/auth/login", authH.Login)
	mux.HandleFunc("/api/auth/forgot-password", authH.ForgotPassword)
	mux.HandleFunc("/api/auth/reset-password", authH.ResetPassword)

	// Product Endpoints
	mux.HandleFunc("/api/products", prodH.SearchOrList)
	mux.HandleFunc("/api/admin/products", admin(prodH.Create))

	// Cart & Payment Endpoints
	mux.HandleFunc("/api/cart", auth(cartH.GetCart))
	mux.HandleFunc("/api/cart/add", auth(cartH.AddItem))
	mux.HandleFunc("/api/cart/item", auth(cartH.UpdateOrRemoveItem))
	mux.HandleFunc("/api/cart/pay", auth(cartH.Pay))

	// User Orders Endpoints
	mux.HandleFunc("/api/orders", auth(orderH.ListUserOrders))

	// Admin Orders & Users Endpoints
	mux.HandleFunc("/api/admin/orders", admin(orderH.AdminListAll))
	mux.HandleFunc("/api/admin/orders/status", admin(orderH.AdminUpdateStatus))
	mux.HandleFunc("/api/admin/orders/create", admin(orderH.AdminCreateOrder))
	mux.HandleFunc("/api/admin/users", admin(userH.ListAll))
	mux.HandleFunc("/api/admin/users/create", admin(userH.CreateUser))
	mux.HandleFunc("/api/admin/users/update", admin(userH.UpdateUser))
	mux.HandleFunc("/api/admin/users/confirm", admin(userH.ConfirmUser))
	mux.HandleFunc("/api/admin/users/delete", admin(userH.DeleteUser))

	return mux
}
