package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func GenerateProductID() string {
	return fmt.Sprintf("PDT-%s", generateRandomString(6))
}

func GenerateCartID() string {
	return fmt.Sprintf("BSK-%s", generateRandomString(6))
}

func GenerateOrderID() string {
	return fmt.Sprintf("CMD-%s", generateRandomString(6))
}

func GenerateConfirmationCode() string {
	return generateRandomString(6)
}
