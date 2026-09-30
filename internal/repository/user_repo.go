package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"ecommerce-cli/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (email, password_hash, role, is_confirmed, confirmation_code) 
	          VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.Exec(query, user.Email, user.PasswordHash, user.Role, user.IsConfirmed, user.ConfirmationCode)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		user.ID = id
	}
	return nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, is_confirmed, confirmation_code, created_at 
	          FROM users WHERE email = ?`
	row := r.db.QueryRow(query, email)

	var u models.User
	var code sql.NullString
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsConfirmed, &code, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if code.Valid {
		u.ConfirmationCode = code.String
	}
	return &u, nil
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, is_confirmed, confirmation_code, created_at 
	          FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var u models.User
	var code sql.NullString
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsConfirmed, &code, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if code.Valid {
		u.ConfirmationCode = code.String
	}
	return &u, nil
}

func (r *UserRepository) ConfirmUser(email, code string) error {
	query := `UPDATE users SET is_confirmed = 1, confirmation_code = '' 
	          WHERE email = ? AND confirmation_code = ?`
	res, err := r.db.Exec(query, email, code)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("invalid code or user not found")
	}
	return nil
}

func (r *UserRepository) UpdatePassword(email, newPasswordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE email = ?`
	res, err := r.db.Exec(query, newPasswordHash, email)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) ListAll() ([]models.User, error) {
	query := `SELECT id, email, role, is_confirmed, created_at FROM users ORDER BY id DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsConfirmed, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
