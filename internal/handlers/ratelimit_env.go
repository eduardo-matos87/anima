package handlers

import (
	"os"
	"strconv"
)

func atoiEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// AtoiEnvInt exportado pra usar no main
func AtoiEnvInt(key string, def int) int { return atoiEnvInt(key, def) }
