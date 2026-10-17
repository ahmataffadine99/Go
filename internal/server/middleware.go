package server

import (
	"context"
	"net/http"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/utils"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func AuthMiddleware(secret string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			userID, role, err := utils.ParseToken(strings.TrimPrefix(authHeader, "Bearer "), secret)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), models.UserIDKey, userID)
			ctx = context.WithValue(ctx, models.RoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func AdminOnlyMiddleware(secret string) Middleware {
	auth := AuthMiddleware(secret)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return auth(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(models.RoleKey).(string)
			if role != "admin" {
				http.Error(w, `{"error":"forbidden: admin access required"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
