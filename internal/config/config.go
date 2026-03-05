package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration. Loaded once at startup; fail fast if invalid.
type Config struct {
	HTTPPort           int
	CORSAllowedOrigins []string
}

// Load reads configuration from environment. Returns error if required values are missing or invalid.
func Load() (*Config, error) {
	portStr := os.Getenv("HTTP_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("HTTP_PORT must be a valid port (1-65535), got %q", portStr)
	}

	allowedOrigins := parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return &Config{HTTPPort: port, CORSAllowedOrigins: allowedOrigins}, nil
}

// parseCORSOrigins splits a comma-separated origins string into a slice.
// Returns ["*"] when the value is empty, allowing all origins by default.
func parseCORSOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}
