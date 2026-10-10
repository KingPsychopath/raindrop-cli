package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestOfflineDiscoveryPreservesWriteBoundaryAndSchema(t *testing.T) {
	for _, test := range []struct {
		args  []string
		count int
		fail  bool
	}{
		{[]string{"tools"}, 26, false},
		{[]string{"tools", "--name", "raindrop_context"}, 1, false},
		{[]string{"tools", "--name", "raindrop_delete"}, 0, true},
		{[]string{"tools", "--allow-write", "--name", "raindrop_delete"}, 1, false},
		{[]string{"tools", "--search", "classify", "--limit", "1"}, 1, false},
		{[]string{"tools", "--search", "zzzzzzzzzz"}, 0, false},
		{[]string{"tools", "--search", "tags", "--name", "raindrop_context"}, 0, true},
		{[]string{"tools", "--limit", "101"}, 0, true},
	} {
		var out bytes.Buffer
		err := (&app{out: &out}).mcp(context.Background(), test.args)
		if (err != nil) != test.fail {
			t.Fatalf("%v: %v", test.args, err)
		}
		if test.fail {
			continue
		}
		var tools []*mcp.Tool
		if err := json.Unmarshal(out.Bytes(), &tools); err != nil || len(tools) != test.count {
			t.Fatalf("%v: tools=%d err=%v", test.args, len(tools), err)
		}
		for _, tool := range tools {
			if tool.InputSchema == nil {
				t.Fatal("discovery lost schema")
			}
		}
		if len(test.args) > 2 && test.args[1] == "--search" && test.args[2] == "classify" && tools[0].Name != "raindrop_classify" {
			t.Fatal("exact name term was not ranked first")
		}
	}
}

func TestMCPCallUsesValidatedSchemaAndReportsToolFailure(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"items":[{"_id":-1,"count":2}]}`))
	}))
	defer server.Close()
	t.Setenv("RAINDROP_TOKEN", "fixture-token")
	t.Setenv("RAINDROP_API_BASE_URL", server.URL)
	var out bytes.Buffer
	a := &app{out: &out}
	if err := a.mcp(context.Background(), []string{"call", "--name", "raindrop_stats"}); err != nil || calls != 1 {
		t.Fatalf("call failed: %v requests=%d", err, calls)
	}
	var result mcp.CallToolResult
	if json.Unmarshal(out.Bytes(), &result) != nil || result.IsError || result.StructuredContent == nil {
		t.Fatal("tool result contract lost")
	}
	input := t.TempDir() + "/arguments.json"
	if err := os.WriteFile(input, []byte(`{"extra":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := a.mcp(context.Background(), []string{"call", "--name", "raindrop_stats", "--input", input}); err == nil || calls != 1 {
		t.Fatal("invalid input reached API or returned success")
	}
}
