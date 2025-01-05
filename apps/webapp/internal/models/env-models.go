package models

type RootConfig struct {
	EnvName         string
	IsProd          bool
	MaxLinksPerUser int
	WebAppConfig    WebAppConfig
}

type WebAppConfig struct {
	Url                string
	Port               int
	GoogleClientId     string
	GoogleClientSecret string
	CookieSecret       string
}
