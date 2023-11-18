package utils

import (
	"math/rand"
)

func GenerateShortKey(storage *map[string]string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6
	ok := true
	var shortKey []byte
	for ok {
		shortKey = make([]byte, keyLength)
		for i := range shortKey {
			shortKey[i] = charset[rand.Intn(len(charset))]
		}
		_, ok = (*storage)[string(shortKey)]
	}

	return string(shortKey)
}
