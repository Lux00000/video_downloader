package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	AuthRequired  bool
	MaxConcurrent int
	RateLimitRPM  int
	YtDlpPath     string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		AuthRequired:  getEnvBool("AUTH_REQUIRED", false),
		MaxConcurrent: getEnvInt("MAX_CONCURRENT", 3),
		RateLimitRPM:  getEnvInt("RATE_LIMIT_RPM", 10),
		YtDlpPath:     getEnv("YTDLP_PATH", "/usr/local/bin/yt-dlp"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

