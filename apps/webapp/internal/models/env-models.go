package models

type RootConfig struct {
	EnvName         string
	IsProd          bool
	MaxLinksPerUser int
	WebAppConfig    WebAppConfig
	Valkey          ValkeyConfig
	AmqpUrl         string
	PostgreSqlDsn   string
}

type WebAppConfig struct {
	Url                string
	Port               int
	GoogleClientId     string
	GoogleClientSecret string
	CookieSecret       string
}

type ValkeyConfig struct {
	Url string
}
