package config

import _ "embed"

//go:embed .env.default
var DefaultEnv string
