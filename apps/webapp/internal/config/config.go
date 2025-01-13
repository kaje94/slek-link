package config

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/kaje94/slek-link/internal/models"
)

var Config = models.RootConfig{
	EnvName:         getEnvValWithFallback("ENV_NAME", "development"),
	IsProd:          getEnvValWithFallback("ENV_NAME", "development") == "production",
	MaxLinksPerUser: getIntEnvValWithFallback("MAX_LINKS_PER_USER", 10),
	Valkey: models.ValkeyConfig{
		Url: getEnvValWithFallback("VALKEY_HOST", ""),
	},
	WebAppConfig: models.WebAppConfig{
		Port:               getIntEnvValWithFallback("WEBAPP_PORT", 8080),
		Url:                getEnvValWithFallback("WEBAPP_URL", "http://localhost:8080"),
		GoogleClientId:     getEnvValWithFallback("GOOGLE_AUTH_CLIENT_ID", ""),
		GoogleClientSecret: getEnvValWithFallback("GOOGLE_AUTH_CLIENT_SECRET", ""),
		CookieSecret:       getEnvValWithFallback("COOKIE_SECRET", "super-secret-key"),
	},
}

func init() {
	isLocalProxy := os.Getenv("IS_LOCAL_PROXY")
	if isLocalProxy != "" {
		boolValue, _ := strconv.ParseBool(isLocalProxy)
		if boolValue {
			slog.Info("Running via a local proxy", "switching port to", 7999)
			Config.WebAppConfig.Port = 7999
		}
	}
}

func getEnvVal(key string) string {
	envVal := os.Getenv(key)
	if envVal == "" {
		log.Fatalf("env for %s is required", key)
	}
	return envVal
}

func getEnvValWithFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnvValWithFallback(key string, fallback int) int {
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
