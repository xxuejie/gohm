package gohm

import (
	"fmt"
)

func toString(v interface{}) string {
	return fmt.Sprint(v)
}

// Works like https://github.com/soveran/nido
func connectKeys(a, b interface{}) string {
	return fmt.Sprintf("%v:%v", a, b)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
