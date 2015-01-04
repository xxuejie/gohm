package gohm

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

func toString(v interface{}) string {
	return fmt.Sprint(v)
}

// Works like https://github.com/soveran/nido
func connectKeys(keys ...interface{}) string {
	strs := make([]string, len(keys))
	for i := range keys {
		strs[i] = toString(keys[i])
	}
	return strings.Join(strs, ":")
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomHexString(n int) (string, error) {
	b, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
