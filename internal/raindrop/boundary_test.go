package raindrop

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedirectAndAPIErrorBoundaries(t *testing.T) {
	escaped := 0
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { escaped++; w.Write([]byte(`{}`)) }))
	defer other.Close()
	mode := "redirect"
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch mode {
		case "redirect":
			http.Redirect(w, r, other.URL, http.StatusFound)
		case "false":
			w.Write([]byte(`{"result":false,"token":"secret-fixture"}`))
		case "rate":
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(429)
			w.Write([]byte(`{"errorMessage":"secret-fixture"}`))
		}
	}))
	defer source.Close()
	api := NewClient(source.URL, "secret-fixture")
	if err := api.Do(context.Background(), "GET", "user", nil, nil, nil); err == nil || escaped != 0 {
		t.Fatal("cross-origin redirect allowed")
	}
	mode = "false"
	if err := api.Do(context.Background(), "POST", "raindrop", nil, map[string]any{}, nil); err == nil || strings.Contains(err.Error(), "secret-fixture") {
		t.Fatal("result:false accepted or token exposed")
	}
	mode = "rate"
	err := api.Do(context.Background(), "GET", "user", nil, nil, nil)
	var typed *APIError
	if !errors.As(err, &typed) || typed.StatusCode != 429 || typed.RetryAfter != "5" || strings.Contains(err.Error(), "secret-fixture") {
		t.Fatalf("bad typed error %v", err)
	}
}
