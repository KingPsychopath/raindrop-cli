package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

func TestJevTypedClassificationAndReviewThreshold(t *testing.T) {
	body := `{"model":"jev-fixture","answers":{"42":{"type":"choice","choice":"docs","confidence":0.7,"probabilities":{"docs":0.9,"needs_review":0.1}}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/systemone" {
			t.Error("wrong Jev endpoint")
		}
		var input map[string]any
		json.NewDecoder(r.Body).Decode(&input)
		state, _ := json.Marshal(input["state"])
		if strings.Contains(string(state), "private-note") {
			t.Error("note unnecessarily sent")
		}
		questions := input["questions"].(map[string]any)
		if questions["42"] == nil {
			t.Error("question missing")
		}
		w.Write([]byte(body))
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "fixture-key")
	items := []raindrop.Raindrop{{ID: 42, Title: "API docs", Note: "private-note"}}
	result, err := client.Classify(context.Background(), "jev-latest", items, map[string]string{"docs": "Technical documentation"}, 0.8, 0.8)
	if err != nil || len(result.Results) != 1 || !result.Results[0].NeedsReview {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	for _, invalid := range []string{`{"answers":{"42":{"type":"choice","choice":"invented","confidence":1,"probabilities":{"docs":1,"needs_review":0}}}}`, `{"answers":{"42":{"type":"choice","choice":"docs","probabilities":{"docs":1,"needs_review":0}}}}`, `{"answers":{"42":{"type":"choice","choice":"docs","confidence":1,"probabilities":{"docs":0.1,"needs_review":0.1}}}}`} {
		body = invalid
		if _, err := client.Classify(context.Background(), "jev-latest", items, map[string]string{"docs": "Docs"}, 0.8, 0.8); err == nil {
			t.Fatal("invalid decision accepted")
		}
	}
}
