package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSecureKey(keyLength int) (string, error) {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	keyString := hex.EncodeToString(key)
	return keyString, nil
}
