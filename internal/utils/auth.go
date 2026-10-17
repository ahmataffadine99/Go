package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	card = strings.ReplaceAll(card, " ", "")
	if len(card) < 12 || len(card) > 19 || !isDigits(card) {
		return false
	}

	sum := 0
	double := false
	for i := len(card) - 1; i >= 0; i-- {
		d := int(card[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func ValidateCVC(cvc string) bool {
	return (len(cvc) == 3 || len(cvc) == 4) && isDigits(cvc)
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s != ""
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
