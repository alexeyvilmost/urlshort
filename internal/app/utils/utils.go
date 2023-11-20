package utils

import (
	"math/rand"
)

func GenerateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}

	return string(shortKey)
}

func GetUniqueShortKey(storage *map[string]string) string {
	ok := true
	var shortKey string
	for ok {
		shortKey = GenerateShortKey()
		_, ok = (*storage)[shortKey]
	}
	return shortKey
}
