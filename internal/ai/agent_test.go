package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func options(api string) AgentOptions {
	return AgentOptions{Model: "fixture-model", API: api, Instructions: "Use tools safely", Prompt: "Find bookmarks", MaxSteps: 3, MaxCalls: 3, MaxOutputTokens: 100, MaxToolBytes: 1024}
}
func fixtureTool() []Tool {
	return []Tool{{Name: "lookup", Description: "Read bookmarks", Schema: map[string]any{"type": "object"}}}
}

func TestResponsesLoopPreservesReasoningAndFeedsTools(t *testing.T) {
	requests := 0
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/responses" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("wrong model route or credential")
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["store"] != false {
			t.Error("model storage was not disabled")
		}
		if requests == 1 {
			w.Write([]byte(`{"status":"completed","output":[{"type":"reasoning","encrypted_content":"opaque","summary":[]},{"type":"function_call","call_id":"call1","name":"lookup","arguments":"{}"}]}`))
			return
		}
		data, _ := json.Marshal(body["input"])
		if !strings.Contains(string(data), "opaque") || !strings.Contains(string(data), "function_call_output") || !strings.Contains(string(data), "bookmark-42") {
			t.Error("continuation lost reasoning or tool result")
		}
		w.Write([]byte(`{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"Found bookmark 42"}]}]}`))
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "secret")
	result, err := client.RunAgent(context.Background(), options("responses"), fixtureTool(), func(ctx context.Context, name string, args json.RawMessage) any {
		calls++
		return map[string]any{"bookmark": "bookmark-42"}
	})
	if err != nil || result.Status != "complete" || result.Steps != 2 || result.ToolCalls != 1 || calls != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestChatLoop(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/chat/completions" {
			t.Error("wrong chat path")
		}
		if requests == 1 {
			w.Write([]byte(`{"choices":[{"finish_reason":"tool_calls","message":{"role":"assistant","content":null,"tool_calls":[{"id":"one","type":"function","function":{"name":"lookup","arguments":"{}"}}]}}]}`))
		} else {
			w.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"Done"}}]}`))
		}
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "")
	result, err := client.RunAgent(context.Background(), options("chat"), fixtureTool(), func(context.Context, string, json.RawMessage) any { return "data" })
	if err != nil || result.Response != "Done" || result.ToolCalls != 1 {
		t.Fatalf("%+v %v", result, err)
	}
}

func TestAgentRejectsInvalidCallsAndLimitsBeforeExecution(t *testing.T) {
	for _, test := range []struct {
		name, body string
		maxCalls   int
	}{{"unknown", `{"status":"completed","output":[{"type":"function_call","call_id":"one","name":"delete_everything","arguments":"{}"}]}`, 3}, {"duplicate", `{"status":"completed","output":[{"type":"function_call","call_id":"one","name":"lookup","arguments":"{}"},{"type":"function_call","call_id":"one","name":"lookup","arguments":"{}"}]}`, 3}, {"budget", `{"status":"completed","output":[{"type":"function_call","call_id":"one","name":"lookup","arguments":"{}"},{"type":"function_call","call_id":"two","name":"lookup","arguments":"{}"}]}`, 1}, {"incomplete", `{"status":"incomplete","output":[{"type":"function_call","call_id":"one","name":"lookup","arguments":"{}"}]}`, 3}} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(test.body)) }))
			defer server.Close()
			client, _ := NewClient(server.URL, "")
			opts := options("responses")
			opts.MaxCalls = test.maxCalls
			executed := 0
			result, err := client.RunAgent(context.Background(), opts, fixtureTool(), func(context.Context, string, json.RawMessage) any { executed++; return nil })
			if err == nil || executed != 0 || result.Status != "stopped" {
				t.Fatalf("executed=%d result=%+v err=%v", executed, result, err)
			}
		})
	}
}

func TestModelHTTPErrorDoesNotEchoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte("secret-account-data"))
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "secret")
	err := client.Post(context.Background(), "responses", map[string]any{}, &map[string]any{})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("error=%v", err)
	}
}

func TestInvalidSchemaInLaterCallBlocksWholeBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"completed","output":[{"type":"function_call","call_id":"first","name":"write","arguments":"{}"},{"type":"function_call","call_id":"second","name":"write","arguments":"{\"invalid\":true}"}]}`))
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "")
	executed := 0
	_, err := client.RunAgent(context.Background(), AgentOptions{Model: "fixture", API: "responses", Prompt: "task", MaxSteps: 1, MaxCalls: 2, MaxOutputTokens: 100, MaxToolBytes: 1024}, []Tool{{Name: "write", Schema: map[string]any{"type": "object"}, Validate: func(raw json.RawMessage) error {
		if string(raw) != "{}" {
			return fmt.Errorf("invalid input")
		}
		return nil
	}}}, func(context.Context, string, json.RawMessage) any { executed++; return map[string]any{} })
	if err == nil || executed != 0 {
		t.Fatalf("schema rejection came after writes: executed=%d err=%v", executed, err)
	}
}
