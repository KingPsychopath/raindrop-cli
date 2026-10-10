package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KingPsychopath/raindrop-cli/internal/organize"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

func TestDirectAgentUsesContextAndCapturesExactValidatedPlan(t *testing.T) {
	item := raindrop.Raindrop{ID: 42, Title: "API docs", Link: "https://fixture.test", Note: "PRIVATE_NOTE", Collection: raindrop.Ref{ID: -1}, Tags: []string{"my tag"}}
	writes := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes++
			t.Error("read-only agent mutated fixture")
		}
		switch r.URL.Path {
		case "/filters/0":
			w.Write([]byte(`{"result":true,"notag":{"count":1},"tags":[]}`))
		case "/raindrops/0":
			json.NewEncoder(w).Encode(map[string]any{"items": []raindrop.Raindrop{item}, "count": 1})
		case "/user":
			w.Write([]byte(`{"user":{"_id":7}}`))
		case "/collections", "/collections/childrens":
			w.Write([]byte(`{"items":[]}`))
		case "/tags/0":
			w.Write([]byte(`{"items":[{"_id":"reference","count":3}]}`))
		case "/raindrop/42":
			json.NewEncoder(w).Encode(map[string]any{"item": item})
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	defer api.Close()
	requests := 0
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/responses" {
			t.Error("wrong protocol")
		}
		var body struct {
			Input []struct {
				Content string `json:"content"`
				Type    string `json:"type"`
				Output  string `json:"output"`
			} `json:"input"`
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		for _, tool := range body.Tools {
			if tool.Name == "raindrop_cleanup_apply" || tool.Name == "raindrop_delete" {
				t.Error("write exposed")
			}
		}
		if requests == 1 {
			initial := body.Input[1].Content
			if strings.Contains(initial, "PRIVATE_NOTE") {
				t.Error("initial model context exposed private note")
			}
			_, raw, ok := strings.Cut(initial, ": ")
			if !ok {
				t.Fatal("missing initial context")
			}
			var context struct {
				Context organize.LibraryContext `json:"context"`
			}
			if err := json.Unmarshal([]byte(raw), &context); err != nil {
				t.Error(err)
			}
			args, _ := json.Marshal(map[string]any{"contextDigest": context.Context.Digest, "suggestions": []map[string]any{{"id": 42, "tags": []string{"reference"}, "reason": "Technical reference", "needsReview": false}}})
			json.NewEncoder(w).Encode(map[string]any{"status": "completed", "output": []any{map[string]any{"type": "function_call", "call_id": "proposal-1", "name": "raindrop_organize_plan", "arguments": string(args)}}})
		} else {
			last := body.Input[len(body.Input)-1]
			if last.Type != "function_call_output" || !strings.Contains(last.Output, "\"action\":\"organize\"") {
				t.Error("model did not receive validated plan")
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "completed", "output": []any{map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": "Review the generated reference-tag plan."}}}}})
		}
	}))
	defer model.Close()
	var out, logs bytes.Buffer
	application := &app{api: raindrop.NewClient(api.URL, "fixture"), out: &out, err: &logs}
	if err := application.agent(context.Background(), []string{"--model", "fixture", "--base-url", model.URL, "--json", "Organize these bookmarks"}); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Status      string            `json:"status"`
		Plans       []organize.Plan   `json:"plans"`
		ToolResults []agentToolResult `json:"toolResults"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.ToolResults) != 1 || result.ToolResults[0].Tool != "raindrop_organize_plan" || result.ToolResults[0].IsError {
		t.Fatalf("missing independent tool results: %s", out.String())
	}
	if requests != 2 || writes != 0 || result.Status != "complete" || len(result.Plans) != 1 || result.Plans[0].Items[0].Note != "PRIVATE_NOTE" || result.Plans[0].Changes[0].Tags[0] != "reference" {
		t.Fatalf("plan round trip failed: %s", out.String())
	}
	if validation, err := organize.Apply(context.Background(), application.api, result.Plans[0], false); err != nil || validation.Status != "validated" {
		t.Fatalf("captured plan is not an exact apply file: %+v %v", validation, err)
	}
}
