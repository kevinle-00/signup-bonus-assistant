package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL        string
	APIAddr            string
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	loadDotEnvIfPresent()

	cfg := Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		APIAddr:            apiAddrFromEnv(),
		CORSAllowedOrigins: splitCSV(getenvDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func apiAddrFromEnv() string {
	if apiAddr := os.Getenv("API_ADDR"); apiAddr != "" {
		return apiAddr
	}

	port := os.Getenv("PORT")
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func getenvDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
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
