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

func TestClassificationIsolatesEachBookmarkBeforeDecision(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var input struct {
			State []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
			} `json:"state"`
			Questions map[string]any `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if len(input.State) != 1 || len(input.Questions) != 1 {
			t.Fatalf("shared bookmark state reached classifier: %+v", input)
		}
		if input.State[0].ID == 42 {
			if input.State[0].Title != "API reference" || input.Questions["42"] == nil {
				t.Fatal("wrong bookmark/question binding")
			}
			w.Write([]byte(`{"answers":{"42":{"type":"choice","choice":"docs","confidence":1,"probabilities":{"docs":1,"needs_review":0}}}}`))
		} else {
			if input.State[0].ID != 43 || input.State[0].Title != "Ambiguous" || input.Questions["43"] == nil {
				t.Fatal("wrong bookmark/question binding")
			}
			w.Write([]byte(`{"answers":{"43":{"type":"choice","choice":"needs_review","confidence":1,"probabilities":{"docs":0,"needs_review":1}}}}`))
		}
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "")
	items := []raindrop.Raindrop{{ID: 42, Title: "API reference"}, {ID: 43, Title: "Ambiguous"}}
	result, err := Classify(context.Background(), client, "local", items, map[string]string{"docs": "Technical documentation"}, 0.8, 0.8)
	if err != nil || calls != 2 || result.Results[0].NeedsReview || !result.Results[1].NeedsReview {
		t.Fatalf("result=%+v calls=%d err=%v", result, calls, err)
	}
	items[1].ID = 42
	if _, err := Classify(context.Background(), client, "local", items, map[string]string{"docs": "Docs"}, 0.8, 0.8); err == nil || calls != 2 {
		t.Fatal("duplicate IDs must fail before contacting the provider")
	}
}
