package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/kaje94/slek-link/internal/models"
	"github.com/kaje94/slek-link/internal/utils"
)

var Config = models.RootConfig{
	EnvName:         utils.GetEnvValWithFallback("ENV_NAME", "development"),
	IsProd:          utils.GetEnvValWithFallback("ENV_NAME", "development") == "production",
	MaxLinksPerUser: utils.GetIntEnvValWithFallback("MAX_LINKS_PER_USER", 10),
	WebAppConfig: models.WebAppConfig{
		Port:               utils.GetIntEnvValWithFallback("WEBAPP_PORT", 8080),
		Url:                utils.GetEnvValWithFallback("WEBAPP_URL", "http://localhost:8080"),
		GoogleClientId:     utils.GetEnvValWithFallback("GOOGLE_AUTH_CLIENT_ID", ""),
		GoogleClientSecret: utils.GetEnvValWithFallback("GOOGLE_AUTH_CLIENT_SECRET", ""),
		CookieSecret:       utils.GetEnvValWithFallback("COOKIE_SECRET", "super-secret-key"),
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
