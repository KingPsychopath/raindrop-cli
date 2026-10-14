package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/machineyearningengineer/raindrop-cli/internal/ai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *app) agent(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	model := fs.String("model", os.Getenv("RDROP_MODEL"), "tool-capable model ID")
	base := fs.String("base-url", os.Getenv("RDROP_MODEL_BASE_URL"), "model API root")
	protocol := fs.String("api", "responses", "responses or chat")
	write := fs.Bool("allow-write", false, "authorize the model to apply writes")
	official := fs.Bool("official", false, "include official hosted Raindrop read tools")
	steps := fs.Int("max-steps", 12, "maximum model requests")
	calls := fs.Int("max-tool-calls", 32, "maximum tool calls")
	tokens := fs.Int("max-output-tokens", 4096, "maximum output tokens per model request")
	timeout := fs.Duration("timeout", 10*time.Minute, "overall time limit")
	jsonOut := fs.Bool("json", false, "print result and exact generated plans")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() == 0 || *model == "" {
		return fmt.Errorf("usage: rdrop agent --model <model> [flags] <task>")
	}
	if *timeout <= 0 {
		return fmt.Errorf("--timeout must be positive")
	}
	if *base == "" {
		*base = "https://api.openai.com/v1"
	}
	key := os.Getenv("RDROP_MODEL_API_KEY")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	parsed, parseErr := url.Parse(*base)
	if parseErr != nil {
		return parseErr
	}
	local := parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1"
	if key == "" && !local {
		return fmt.Errorf("set RDROP_MODEL_API_KEY (or OPENAI_API_KEY); local loopback models may run without a key")
	}
	client, err := ai.NewClient(*base, key)
	if err != nil {
		return err
	}
	engine := &toolEngine{api: a.api, write: *write}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	if *official {
		if err := engine.enableOfficial(ctx); err != nil {
			return err
		}
		defer engine.official.session.Close()
	}
	tools := []ai.Tool{}
	for _, tool := range engine.tools() {
		tools = append(tools, ai.Tool{Name: tool.Name, Description: tool.Description, Schema: tool.InputSchema, Validate: func(input json.RawMessage) error { return validateToolInput(tool, input) }})
	}
	initial := engine.call(ctx, "raindrop_cleanup_report", json.RawMessage(`{}`))
	if initial.IsError {
		return fmt.Errorf("initial cleanup report failed: %s", initial.Content[0].(*mcp.TextContent).Text)
	}
	library := engine.call(ctx, "raindrop_context", json.RawMessage(`{}`))
	if library.IsError {
		return fmt.Errorf("initial context failed: %s", library.Content[0].(*mcp.TextContent).Text)
	}
	data, _ := json.Marshal(map[string]any{"report": initial.StructuredContent, "context": library.StructuredContent})
	plans := []json.RawMessage{}
	audit := []agentToolResult{}
	result, runErr := client.RunAgent(ctx, ai.AgentOptions{Model: *model, API: *protocol, Instructions: organizeInstructions + " Return a clear response. Never invent API results or rewrite snapshot plans. Read-only mode can generate plans but cannot apply them. Jev confidence is not proof of correctness. Use pagination and don't claim a whole-library scan from one page.", Prompt: strings.Join(fs.Args(), " "), Initial: string(data), MaxSteps: *steps, MaxCalls: *calls, MaxOutputTokens: *tokens, MaxToolBytes: 256 << 10}, tools, func(ctx context.Context, name string, input json.RawMessage) any {
		output := engine.call(ctx, name, input)
		entry := agentToolResult{Tool: name, IsError: output.IsError, Output: output.StructuredContent}
		if entry.Output == nil {
			entry.Output = output.Content
		}
		if raw, _ := json.Marshal(entry.Output); len(raw) > 256<<10 {
			entry.Output = nil
			entry.Truncated = true
		}
		audit = append(audit, entry)
		if !output.IsError && strings.HasSuffix(name, "_preview") {
			plan, _ := json.Marshal(output.StructuredContent)
			plans = append(plans, plan)
		}
		if !output.IsError && name == "raindrop_organize_plan" {
			raw, _ := json.Marshal(output.StructuredContent)
			var parsed struct {
				Plan json.RawMessage `json:"plan"`
			}
			if json.Unmarshal(raw, &parsed) == nil && len(parsed.Plan) > 0 {
				plans = append(plans, parsed.Plan)
			}
		}
		return output
	})
	if *jsonOut || runErr != nil {
		if err := a.print(struct {
			ai.AgentResult
			Plans       []json.RawMessage `json:"plans"`
			ToolResults []agentToolResult `json:"toolResults"`
		}{result, plans, audit}, nil); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(a.out, result.Response)
		if len(plans) > 0 {
			fmt.Fprintln(a.err, "rdrop: use --json to capture exact snapshot plans; model prose is not an apply file")
		}
	}
	return runErr
}

type agentToolResult struct {
	Tool      string `json:"tool"`
	IsError   bool   `json:"isError"`
	Output    any    `json:"output,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}
