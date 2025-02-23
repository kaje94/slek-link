package utils

type sessionKeys string

const (
	DB_CONTEXT_KEY      sessionKeys = "DB-CONTEXT"
	VALKEY_CONTEXT_KEY  sessionKeys = "VALKEY-CONTEXT-KEY"
	SESSION_CONTEXT_KEY sessionKeys = "SESSION-CONTEXT"
	LayoutHuman                     = "Jan 02, 2006"
)
