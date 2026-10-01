package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
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
	if len(expiry) != 5 || expiry[2] != '/' {
		return false
	}

	monthStr := expiry[:2]
	yearStr := "20" + expiry[3:]

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return false
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return false
	}

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	if year < currentYear {
		return false
	}
	if year == currentYear && month < currentMonth {
		return false
	}

	return true
}

var ErrInvalidInput = errors.New("invalid input data")
