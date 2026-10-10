package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolate(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	for _, key := range []string{"XDG_CONFIG_HOME", "RAINDROP_TOKEN", "RAINDROP_TOKEN_FILE", "RAINDROP_CONFIG_FILE", "RAINDROP_API_BASE_URL", "RAINDROP_MCP_ENDPOINT"} {
		t.Setenv(key, "")
	}
}

func TestTokenPrecedenceAndDiagnostics(t *testing.T) {
	isolate(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "env")
	if err := os.WriteFile(file, []byte("export RAINDROP_TOKEN='file-secret'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RAINDROP_CONFIG_FILE", file)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "file-secret" || cfg.TokenSource != file {
		t.Fatalf("wrong file source")
	}
	tokenFile := filepath.Join(dir, "token")
	os.WriteFile(tokenFile, []byte("token-secret\n"), 0600)
	t.Setenv("RAINDROP_TOKEN_FILE", tokenFile)
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "token-secret" {
		t.Fatal("token file did not take precedence")
	}
	t.Setenv("RAINDROP_TOKEN", "environment-secret")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(cfg.Status())
	if cfg.Token != "environment-secret" || strings.Contains(string(data), "secret") {
		t.Fatal("diagnostics exposed token or precedence failed")
	}
}

func TestNativeAndLegacyPaths(t *testing.T) {
	isolate(t)
	files := EnvFiles()
	if len(files) != 2 || !strings.HasSuffix(files[0], filepath.Join("rdrop", "env")) {
		t.Fatalf("paths=%v", files)
	}
	os.MkdirAll(filepath.Dir(files[1]), 0700)
	os.WriteFile(files[1], []byte("RAINDROP_TOKEN=legacy\n"), 0600)
	cfg, err := Load()
	if err != nil || cfg.Token != "legacy" {
		t.Fatalf("legacy fallback: %v", err)
	}
	os.MkdirAll(filepath.Dir(files[0]), 0700)
	os.WriteFile(files[0], []byte("RAINDROP_TOKEN=native\n"), 0600)
	cfg, err = Load()
	if err != nil || cfg.Token != "native" {
		t.Fatalf("native precedence: %v", err)
	}
}

func TestMissingAndInvalidConfiguration(t *testing.T) {
	isolate(t)
	if _, err := Read(); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("expected missing token error")
	}
	t.Setenv("RAINDROP_TOKEN_FILE", filepath.Join(t.TempDir(), "missing"))
	if _, err := Read(); err == nil {
		t.Fatal("explicit missing file was ignored")
	}
	for _, value := range []string{"http://example.com", "https://user:pass@example.com", "https://example.com?token=x", "file:///tmp/x"} {
		if ValidateURL(value) == nil {
			t.Fatalf("accepted %s", value)
		}
	}
	for _, value := range []string{"https://example.com/rest/v1", "http://127.0.0.1:1234/v1", "http://[::1]:1234/v1"} {
		if err := ValidateURL(value); err != nil {
			t.Fatal(err)
		}
	}
}
