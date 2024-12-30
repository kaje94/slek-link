package config

type rootConfig struct {
	EnvName         string
	IsProd          bool
	MaxLinksPerUser int
	WebAppConfig    webAppConfig
}

type webAppConfig struct {
	Url                string
	Port               int
	GoogleClientId     string
	GoogleClientSecret string
	CookieSecret       string
}
