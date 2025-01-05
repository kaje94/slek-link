package utils

import (
	"log"
	"os"
	"strconv"
)

func GetEnvVal(key string) string {
	envVal := os.Getenv(key)
	if envVal == "" {
		log.Fatalf("env for %s is required", key)
	}
	return envVal
}

func GetEnvValWithFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetIntEnvValWithFallback(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if value == "" {
			log.Fatalf("env for %s is required", key)
		}
		envValInt, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("failed to parse %s into an int", key)
		}
		return envValInt
	}
	return fallback
}
