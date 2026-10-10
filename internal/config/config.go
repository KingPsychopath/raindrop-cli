package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const DefaultEndpoint = "https://api.raindrop.io/rest/v2/ai/mcp"
const DefaultAPIBaseURL = "https://api.raindrop.io/rest/v1"

type Config struct {
	Endpoint    string
	APIBaseURL  string
	Token       string
	TokenSource string
}

func Load() (Config, error) {
	cfg, err := Read()
	if err != nil {
		return cfg, err
	}
	if cfg.Token == "" {
		return cfg, errors.New("missing RAINDROP_TOKEN; set it in the environment, RAINDROP_TOKEN_FILE, or the rdrop config env file (see rdrop config paths)")
	}
	return cfg, nil
}

// Read resolves configuration without requiring a token. Never serialize Config:
// use Status for diagnostics so credentials stay out of output.
func Read() (Config, error) {
	cfg := Config{
		Endpoint:   getenvDefault("RAINDROP_MCP_ENDPOINT", DefaultEndpoint),
		APIBaseURL: getenvDefault("RAINDROP_API_BASE_URL", DefaultAPIBaseURL),
		Token:      os.Getenv("RAINDROP_TOKEN"),
	}
	for name, endpoint := range map[string]string{"RAINDROP_API_BASE_URL": cfg.APIBaseURL, "RAINDROP_MCP_ENDPOINT": cfg.Endpoint} {
		if err := ValidateURL(endpoint); err != nil {
			return cfg, fmt.Errorf("%s: %w", name, err)
		}
	}
	if cfg.Token != "" {
		cfg.TokenSource = "environment"
	} else if file := os.Getenv("RAINDROP_TOKEN_FILE"); file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return cfg, fmt.Errorf("read RAINDROP_TOKEN_FILE: %w", err)
		}
		cfg.Token = strings.TrimSpace(string(data))
		cfg.TokenSource = file
		if cfg.Token == "" {
			return cfg, errors.New("RAINDROP_TOKEN_FILE is empty")
		}
	} else {
		for _, file := range EnvFiles() {
			token, err := readTokenFromEnvFile(file)
			if err != nil {
				return cfg, err
			}
			if token != "" {
				cfg.Token, cfg.TokenSource = token, file
				break
			}
		}
	}
	if strings.ContainsAny(cfg.Token, "\r\n") {
		return cfg, errors.New("Raindrop token must be a single line")
	}
	return cfg, nil
}

// ValidateURL permits secure remote endpoints and HTTP loopback fixtures.
func ValidateURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("expected an absolute HTTP(S) URL without credentials, query, or fragment")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1") {
		return nil
	}
	return errors.New("remote endpoints require HTTPS; HTTP is allowed only on loopback")
}

type Status struct {
	APIBaseURL      string   `json:"apiBaseURL"`
	MCPEndpoint     string   `json:"mcpEndpoint"`
	EnvFiles        []string `json:"envFiles"`
	TokenConfigured bool     `json:"tokenConfigured"`
	TokenSource     string   `json:"tokenSource,omitempty"`
}

func (c Config) Status() Status {
	return Status{c.APIBaseURL, c.Endpoint, EnvFiles(), c.Token != "", c.TokenSource}
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
	if dir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, "rdrop", "env")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "rdrop", "env")
}

func EnvFiles() []string {
	if file := os.Getenv("RAINDROP_CONFIG_FILE"); file != "" {
		return []string{file}
	}
	files := []string{defaultEnvFile()}
	if home, err := os.UserHomeDir(); err == nil {
		files = append(files, filepath.Join(home, ".config", "openclaw", "gateway.env"))
	}
	return files
}

func readTokenFromEnvFile(file string) (string, error) {
	if file == "" {
		return "", nil
	}
	f, err := os.Open(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && os.Getenv("RAINDROP_CONFIG_FILE") == "" {
			return "", nil
		}
		return "", fmt.Errorf("read config env file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "export ")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "RAINDROP_TOKEN" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"')) {
			value = value[1 : len(value)-1]
		}
		return value, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read config env file: %w", err)
	}
	return "", nil
}
