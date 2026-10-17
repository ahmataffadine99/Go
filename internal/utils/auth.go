package utils

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	pbkdf2Iterations = 100000
	tokenDuration    = 24 * time.Hour
)

var ErrInvalidToken = errors.New("invalid token")

func HashPassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, 32)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(key)
}

func VerifyPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, len(expected))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(key, expected) == 1
}

func GenerateToken(userID int64, role, secret string) string {
	expires := time.Now().Add(tokenDuration).Unix()
	payload := fmt.Sprintf("%d|%s|%d", userID, role, expires)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return encoded + "." + sign(encoded, secret)
}

func ParseToken(token, secret string) (int64, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, "", ErrInvalidToken
	}
	if !hmac.Equal([]byte(parts[1]), []byte(sign(parts[0], secret))) {
		return 0, "", ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, "", ErrInvalidToken
	}
	fields := strings.Split(string(raw), "|")
	if len(fields) != 3 {
		return 0, "", ErrInvalidToken
	}

	userID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, "", ErrInvalidToken
	}
	expires, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return 0, "", ErrInvalidToken
	}
	return userID, fields[1], nil
}

func sign(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
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
