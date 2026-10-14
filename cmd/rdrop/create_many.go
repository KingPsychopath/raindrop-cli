package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/machineyearningengineer/raindrop-cli/internal/organize"
	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *app) batchAdd(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("batch add", flag.ContinueOnError)
	file := fs.String("file", "", "JSON array of 1..100 new bookmarks")
	planFile := fs.String("plan", "", "exact reviewed creation plan")
	dry := fs.Bool("dry-run", false, "plan or validate without creating")
	yes := fs.Bool("yes", false, "apply the reviewed creation plan")
	fs.Bool("json", true, "always print structured results")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 0 || (*file == "") == (*planFile == "") || *dry == *yes || (*file != "" && !*dry) {
		return fmt.Errorf("use batch add --file items.json --dry-run or --plan reviewed-plan.json --dry-run|--yes")
	}
	if *file != "" {
		var items []raindrop.CreateItem
		if err := readJSONFile(*file, &items); err != nil {
			return err
		}
		return a.print(organize.NewCreatePlan(ctx, a.api, items))
	}
	var plan organize.CreatePlan
	if err := readJSONFile(*planFile, &plan); err != nil {
		return err
	}
	result, err := organize.ApplyCreatePlan(ctx, a.api, plan, *yes)
	if printErr := a.print(result, nil); printErr != nil {
		return printErr
	}
	return err
}

func createTools(write bool) []*mcp.Tool {
	item := map[string]any{"type": "object", "properties": map[string]any{
		"link": map[string]any{"type": "string", "minLength": 1}, "title": map[string]any{"type": "string"}, "note": map[string]any{"type": "string"}, "excerpt": map[string]any{"type": "string"},
		"tags":       map[string]any{"type": "array", "items": map[string]any{"type": "string", "minLength": 1}},
		"collection": map[string]any{"type": "object", "properties": map[string]any{"$id": map[string]any{"type": "integer"}}, "required": []string{"$id"}, "additionalProperties": false},
		"important":  map[string]any{"type": "boolean"}, "pleaseParse": map[string]any{"type": "object", "additionalProperties": false},
	}, "required": []string{"link"}, "additionalProperties": false}
	tools := []*mcp.Tool{{Name: "raindrop_add_batch_preview", Description: "Preview 1..100 HTTP(S) bookmark creations as an exact account-bound plan. Validates destination collections; never creates. Review URLs, tags and notes before apply.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"items": map[string]any{"type": "array", "items": item, "minItems": 1, "maxItems": 100}}, "required": []string{"items"}, "additionalProperties": false}, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPointer(false)}}}
	if write {
		tools = append(tools, &mcp.Tool{Name: "raindrop_add_batch_apply", Description: "Create bookmarks from an exact reviewed account-bound creation plan. Requires explicit authorization; errors may leave an uncertain or partial outcome. Inspect saved URLs before retrying.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"plan": map[string]any{"type": "object"}, "confirm": map[string]any{"type": "boolean", "const": true}}, "required": []string{"plan", "confirm"}, "additionalProperties": false}, Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPointer(false)}})
	}
	return tools
}

func (e *toolEngine) createCall(ctx context.Context, name string, input json.RawMessage) *mcp.CallToolResult {
	if name == "raindrop_add_batch_preview" {
		var value struct {
			Items []raindrop.CreateItem `json:"items"`
		}
		if err := json.Unmarshal(input, &value); err != nil {
			return toolError(err)
		}
		plan, err := organize.NewCreatePlan(ctx, e.api, value.Items)
		if err != nil {
			return toolError(err)
		}
		return toolJSON(plan)
	}
	var value struct {
		Plan    organize.CreatePlan `json:"plan"`
		Confirm bool                `json:"confirm"`
	}
	if err := json.Unmarshal(input, &value); err != nil {
		return toolError(err)
	}
	result, err := organize.ApplyCreatePlan(ctx, e.api, value.Plan, value.Confirm)
	output := toolJSON(result)
	if err != nil {
		output.IsError = true
		output.Content = append(output.Content, &mcp.TextContent{Text: err.Error()})
	}
	return output
}
