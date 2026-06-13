package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Settings holds every runtime configuration value the server needs.
// All fields are populated from environment variables; .env is loaded
// in development as a convenience and is silently ignored if missing.
type Settings struct {
	Port int

	DatabaseURL string

	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	JWTExpiryMin int

	SuperUserUsername string
	SuperUserPassword string
	SuperUserEmail    string
}

// Load reads .env (if present) then resolves all required env vars into a
// Settings struct. Missing required vars produce a clear error listing
// every missing key in one go.
func Load() (Settings, error) {
	_ = godotenv.Load()

	var missing []string
	get := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	s := Settings{
		DatabaseURL:       get("DATABASE_URL"),
		JWTSecret:         get("JWT_SECRET"),
		JWTIssuer:         get("JWT_ISSUER"),
		JWTAudience:       get("JWT_AUDIENCE"),
		SuperUserUsername: get("SUPERUSER_USERNAME"),
		SuperUserPassword: get("SUPERUSER_PASSWORD"),
		SuperUserEmail:    get("SUPERUSER_EMAIL"),
	}

	s.Port = intEnv("PORT", 8080)
	s.JWTExpiryMin = intEnv("JWT_EXPIRY_MIN", 60)

	if len(missing) > 0 {
		return Settings{}, fmt.Errorf("missing required env vars: %v", missing)
	}
	if len(s.JWTSecret) < 32 {
		return Settings{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	return s, nil
}

func intEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
