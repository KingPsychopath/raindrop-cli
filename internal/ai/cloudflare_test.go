package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
)

func TestCloudflareDecisionEnvelopeAndBoundaries(t *testing.T) {
	account := strings.Repeat("a", 32)
	body := `{"success":true,"result":{"model":"clef-flash","answers":{"42":{"type":"choice","choice":"docs","confidence":0.7,"probabilities":{"docs":0.9,"needs_review":0.1}}}}}`
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/accounts/"+account+"/ai/run/@cf/cloudflare/clef-flash" || r.Header.Get("Authorization") != "Bearer fixture-key" {
			t.Fatal("wrong Workers AI request or credentials")
		}
		var request struct {
			Model string           `json:"model"`
			State []map[string]any `json:"state"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Model != "clef-flash" || len(request.State) != 1 || request.State[0]["note"] != nil {
			t.Fatalf("invalid isolated decision request: %+v %v", request, err)
		}
		w.Write([]byte(body))
	}))
	defer server.Close()
	provider, err := NewCloudflareProvider(server.URL, "fixture-key", account, "clef-flash")
	if err != nil {
		t.Fatal(err)
	}
	items := []raindrop.Raindrop{{ID: 42, Title: "API reference", Note: "private"}}
	result, err := Classify(context.Background(), provider, "clef-flash", items, map[string]string{"docs": "Technical reference"}, 0.8, 0.8)
	if err != nil || len(result.Results) != 1 || !result.Results[0].NeedsReview {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	for _, invalid := range []string{`{"success":false,"result":{"answers":{}},"errors":[{"message":"secret"}]}`, `{"success":true,"result":null}`} {
		body = invalid
		if _, err := Classify(context.Background(), provider, "clef-flash", items, map[string]string{"docs": "Reference"}, 0.8, 0.8); err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("unsuccessful envelope accepted or upstream data exposed")
		}
	}
	for _, invalid := range []struct{ account, model string }{{"../other", "clef-flash"}, {account, "../other"}} {
		if _, err := NewCloudflareProvider(server.URL, "fixture-key", invalid.account, invalid.model); err == nil {
			t.Fatal("invalid provider path accepted")
		}
	}
	if calls != 3 {
		t.Fatalf("unexpected request count %d", calls)
	}
}
