package models

type User struct {
	ID               int64  `json:"id"`
	Email            string `json:"email"`
	PasswordHash     string `json:"-"`
	Role             string `json:"role"`
	IsConfirmed      bool   `json:"is_confirmed"`
	ConfirmationCode string `json:"-"`
	CreatedAt        string `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ConfirmRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}
