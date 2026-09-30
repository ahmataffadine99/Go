package server

import (
	"database/sql"
	"net/http"

	"ecommerce-cli/internal/repository"
	"ecommerce-cli/internal/server/handlers"
)

func NewRouter(db *sql.DB, driverName string) http.Handler {
	mux := http.NewServeMux()

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	productRepo.SetDriverName(driverName)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	authH := handlers.NewAuthHandler(userRepo)
	prodH := handlers.NewProductHandler(productRepo)
	cartH := handlers.NewCartHandler(cartRepo, orderRepo)
	orderH := handlers.NewOrderHandler(orderRepo, userRepo)
	userH := handlers.NewUserHandler(userRepo)

	// Public Auth Endpoints
	mux.HandleFunc("/api/auth/register", authH.Register)
	mux.HandleFunc("/api/auth/confirm", authH.Confirm)
	mux.HandleFunc("/api/auth/login", authH.Login)
	mux.HandleFunc("/api/auth/reset-password", authH.ResetPassword)

	// Product Endpoints
	mux.HandleFunc("/api/products", prodH.SearchOrList)
	mux.HandleFunc("/api/admin/products", AdminOnlyMiddleware(prodH.Create))

	// Cart & Payment Endpoints
	mux.HandleFunc("/api/cart", AuthMiddleware(cartH.GetCart))
	mux.HandleFunc("/api/cart/add", AuthMiddleware(cartH.AddItem))
	mux.HandleFunc("/api/cart/item", AuthMiddleware(cartH.UpdateOrRemoveItem))
	mux.HandleFunc("/api/cart/pay", AuthMiddleware(cartH.Pay))

	// User Orders Endpoints
	mux.HandleFunc("/api/orders", AuthMiddleware(orderH.ListUserOrders))

	// Admin Orders & Users Endpoints
	mux.HandleFunc("/api/admin/orders", AdminOnlyMiddleware(orderH.AdminListAll))
	mux.HandleFunc("/api/admin/orders/status", AdminOnlyMiddleware(orderH.AdminUpdateStatus))
	mux.HandleFunc("/api/admin/users", AdminOnlyMiddleware(userH.ListAll))
	mux.HandleFunc("/api/admin/users/create", AdminOnlyMiddleware(userH.CreateUser))
	mux.HandleFunc("/api/admin/users/delete", AdminOnlyMiddleware(userH.DeleteUser))

	return mux
}
