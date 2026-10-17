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

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type AdminUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
