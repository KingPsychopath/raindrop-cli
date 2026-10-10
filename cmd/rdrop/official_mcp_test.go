package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestOfficialBridgeReadsSchemasAndExcludesUpstreamWrites(t *testing.T) {
	var calls atomic.Int32
	upstream := mcp.NewServer(&mcp.Implementation{Name: "official-fixture", Version: "1"}, nil)
	schema := map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}, "required": []string{"query"}, "additionalProperties": false}
	for _, name := range []string{"find_bookmarks", "delete_bookmarks", "unrecognized_read"} {
		upstream.AddTool(&mcp.Tool{Name: name, InputSchema: schema, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			calls.Add(1)
			return toolJSON(map[string]any{"items": []any{}, "source": "official-fixture"}), nil
		})
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return upstream }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Error("missing bound credential")
			http.Error(w, "unauthorized", 401)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	bridge, err := connectOfficial(ctx, server.URL, "fixture-token")
	if err != nil {
		t.Fatal(err)
	}
	defer bridge.session.Close()
	engine := &toolEngine{official: bridge, write: true}
	if len(bridge.tools) != 1 || bridge.tools[0].Name != "raindrop_official_find_bookmarks" {
		t.Fatalf("unsafe catalog %+v", bridge.tools)
	}
	for _, call := range []struct{ name, input string }{{"raindrop_official_delete_bookmarks", `{"query":"x"}`}, {"raindrop_official_unrecognized_read", `{"query":"x"}`}, {"raindrop_official_find_bookmarks", `{"extra":true}`}} {
		if result := engine.call(ctx, call.name, json.RawMessage(call.input)); !result.IsError {
			t.Fatal("unsafe remote call accepted")
		}
	}
	result := engine.call(ctx, "raindrop_official_find_bookmarks", json.RawMessage(`{"query":"design references"}`))
	if result.IsError || result.StructuredContent == nil || calls.Load() != 1 {
		t.Fatalf("result=%+v calls=%d", result, calls.Load())
	}
}

func TestOfficialBridgeNeverForwardsCredentialsOnRedirect(t *testing.T) {
	var reached atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached.Add(1) }))
	defer other.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, other.URL, 307) }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := connectOfficial(ctx, server.URL, "fixture-token"); err == nil || strings.Contains(err.Error(), "fixture-token") || reached.Load() != 0 {
		t.Fatalf("unsafe redirect reached=%d err=%v", reached.Load(), err)
	}
}
