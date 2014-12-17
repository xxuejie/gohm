package gohm

import (
	"fmt"
	"strings"
)

func toString(v interface{}) string {
	return fmt.Sprint(v)
}

// Works like https://github.com/soveran/nido
func connectKeys(keys ... interface{}) string {
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
