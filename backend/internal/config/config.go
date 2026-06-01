package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	APIAddr     string
}

func Load() (Config, error) {
	loadDotEnvIfPresent()

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		APIAddr:     getenvDefault("API_ADDR", ":8080"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func getenvDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadDotEnvIfPresent() {
	for _, path := range []string{".env", "../.env"} {
		if loadDotEnv(path) == nil {
			return
		}
	}
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}

	return scanner.Err()
}
