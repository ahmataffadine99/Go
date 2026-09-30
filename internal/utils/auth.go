package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

func GenerateSimpleToken(userID int64, role string) string {
	return fmt.Sprintf("token_%d_%s_%d", userID, role, time.Now().Unix())
}

func ValidateCardNumber(card string) bool {
	return len(card) >= 12 && len(card) <= 19
}

func ValidateCVC(cvc string) bool {
	return len(cvc) == 3 || len(cvc) == 4
}

func ValidateExpiry(expiry string) bool {
	return len(expiry) == 5 && expiry[2] == '/'
}

var ErrInvalidInput = errors.New("invalid input data")
