package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const DefaultEndpoint = "https://api.raindrop.io/rest/v2/ai/mcp"
const DefaultAPIBaseURL = "https://api.raindrop.io/rest/v1"

type Config struct {
	Endpoint   string
	APIBaseURL string
	Token      string
}

func Load() (Config, error) {
	cfg := Config{
		Endpoint:   getenvDefault("RAINDROP_MCP_ENDPOINT", DefaultEndpoint),
		APIBaseURL: getenvDefault("RAINDROP_API_BASE_URL", DefaultAPIBaseURL),
		Token:      os.Getenv("RAINDROP_TOKEN"),
	}
	if cfg.Token == "" {
		cfg.Token = readTokenFromEnvFile(defaultEnvFile())
	}
	if cfg.Token == "" {
		return cfg, errors.New("missing RAINDROP_TOKEN; set it in the environment or ~/.config/openclaw/gateway.env")
	}
	return cfg, nil
}

func DefaultEnvFileForDisplay() string {
	return defaultEnvFile()
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func defaultEnvFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "openclaw", "gateway.env")
}

func readTokenFromEnvFile(file string) string {
	if file == "" {
		return ""
	}
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "RAINDROP_TOKEN" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}
