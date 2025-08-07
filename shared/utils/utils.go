package utils

import (
	"os"
)

func GetEnvOrDefault(env string, def string) string {
	envTry := os.Getenv(env)
	if envTry == "" {
		return def
	}
	return envTry
}
