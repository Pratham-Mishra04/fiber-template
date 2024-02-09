package config

import (
	"time"
)

const (
	ACCESS_TOKEN_TTL  = 15 * time.Minute
	REFRESH_TOKEN_TTL = 14 * 24 * time.Hour
	SIGN_UP_TOKEN_TTL = 1 * time.Minute
	LOGIN_TOKEN_TTL   = 30 * time.Second
)
