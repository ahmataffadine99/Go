package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-cli/internal/models"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")
		tokenParts := strings.Split(rawToken, "_")
		if len(tokenParts) < 3 || tokenParts[0] != "token" {
			http.Error(w, `{"error":"invalid token format"}`, http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(tokenParts[1], 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid token payload"}`, http.StatusUnauthorized)
			return
		}

		role := tokenParts[2]

		ctx := context.WithValue(r.Context(), models.UserIDKey, userID)
		ctx = context.WithValue(ctx, models.RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func AdminOnlyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(models.RoleKey).(string)
		if role != "admin" {
			http.Error(w, `{"error":"forbidden: admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
