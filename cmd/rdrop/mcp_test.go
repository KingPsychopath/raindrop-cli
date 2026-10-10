package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPProcess(t *testing.T) {
	if os.Getenv("RDROP_TEST_MCP_PROCESS") != "1" {
		return
	}
	if err := run(context.Background(), []string{"mcp", "serve"}); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestMCPStdioHandshakeAndReadTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/user/stats" || r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Errorf("unexpected API request %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"result":true,"items":[],"meta":{"_id":7}}`))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.Command(os.Args[0], "-test.run=^TestMCPProcess$")
	command.Env = append(os.Environ(), "RDROP_TEST_MCP_PROCESS=1", "RAINDROP_TOKEN=fixture-token", "RAINDROP_API_BASE_URL="+server.URL)
	client := mcp.NewClient(&mcp.Implementation{Name: "integration-fixture", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: command}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatalf("write tool exposed by default: %s", tool.Name)
		}
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "raindrop_stats", Arguments: map[string]any{}})
	if err != nil || result.IsError || result.StructuredContent == nil {
		t.Fatalf("call result=%+v error=%v", result, err)
	}
	resources, err := session.ListResources(ctx, nil)
	if err != nil || len(resources.Resources) != 1 {
		t.Fatalf("resources=%v error=%v", resources, err)
	}
	prompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{Name: "organize_library"})
	if err != nil || len(prompt.Messages) != 1 {
		t.Fatalf("prompt=%v error=%v", prompt, err)
	}
}

func TestToolWritesAndInputCannotBypassBoundaries(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++; w.Write([]byte(`{"result":true}`)) }))
	defer server.Close()
	engine := &toolEngine{api: raindrop.NewClient(server.URL, "fixture")}
	for _, call := range []struct{ name, args string }{{"raindrop_delete", `{"id":1,"confirm":true}`}, {"raindrop_list", `{"perPage":51}`}, {"raindrop_get", `{"id":1,"extra":"--yes"}`}, {"raindrop_list", `{"page":-1}`}} {
		if result := engine.call(context.Background(), call.name, json.RawMessage(call.args)); !result.IsError {
			t.Fatalf("unsafe input accepted: %+v", call)
		}
	}
	engine.write = true
	if result := engine.call(context.Background(), "raindrop_delete", json.RawMessage(`{"id":1,"confirm":false}`)); !result.IsError {
		t.Fatal("write without confirm accepted")
	}
	if requests != 0 {
		t.Fatal("rejected calls reached API")
	}
}

func TestTokenlessDiscoveryAndConnectionOutputs(t *testing.T) {
	t.Setenv("RAINDROP_TOKEN", "")
	for _, args := range [][]string{{"mcp", "tools"}, {"connect", "mcp"}, {"connect", "acp"}, {"connect", "executor"}, {"config", "paths"}, {"help", "cleanup"}} {
		var out bytes.Buffer
		if err := (&app{out: &out, err: &out}).run(context.Background(), args); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if out.Len() == 0 {
			t.Fatalf("no output for %v", args)
		}
	}
}

func TestReportSearchQuotingAndScopedTagPlans(t *testing.T) {
	report := cleanupReport(10, "$(touch injected)", raindrop.Filters{Tags: []raindrop.Tag{{ID: "Docs", Count: 2}}})
	if len(report.TagPlans) != 0 || !strings.Contains(report.Commands["filters"], "'$(touch injected)'") {
		t.Fatalf("unsafe report=%+v", report)
	}
}
