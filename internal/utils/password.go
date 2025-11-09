package utils

import (
	"crypto/rand"
	"math/big"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GeneratePassword создаёт случайный пароль заданной длины.
func GeneratePassword(length int) (string, error) {
	pass := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		pass[i] = letters[n.Int64()]
	}
	return string(pass), nil
}
