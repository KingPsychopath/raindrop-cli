package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/machineyearningengineer/raindrop-cli/internal/ai"
	"github.com/machineyearningengineer/raindrop-cli/internal/cli"
	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type classifyInput struct {
	IDs            []int64           `json:"ids"`
	Categories     map[string]string `json:"categories"`
	MinConfidence  *float64          `json:"minConfidence"`
	MinProbability *float64          `json:"minProbability"`
}

func classifyTool() *mcp.Tool {
	return &mcp.Tool{Name: "raindrop_classify", Description: "Classify 1..25 bookmarks into fixed caller-defined categories using the configured System One HTTP or local command provider (Jev, Laya, Cloudflare Clef, or a compatible adapter). Sends metadata to that provider; no private notes. Validates distributions and returns needsReview. Never mutates.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "minItems": 1, "maxItems": 25, "uniqueItems": true}, "categories": map[string]any{"type": "object", "minProperties": 1, "maxProperties": 254, "additionalProperties": map[string]any{"type": "string", "minLength": 1}}, "minConfidence": map[string]any{"type": "number", "minimum": 0, "maximum": 1}, "minProbability": map[string]any{"type": "number", "minimum": 0, "maximum": 1}}, "required": []string{"ids", "categories"}, "additionalProperties": false}, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPointer(false)}}
}

func (a *app) classifyItems(ctx context.Context, input classifyInput) (ai.ClassificationResult, error) {
	if err := validateIDs(input.IDs); err != nil {
		return ai.ClassificationResult{}, err
	}
	if len(input.IDs) == 0 || len(input.IDs) > 25 {
		return ai.ClassificationResult{}, fmt.Errorf("select 1..25 IDs")
	}
	provider, model, err := decisionProvider()
	if err != nil {
		return ai.ClassificationResult{}, err
	}
	items := []raindrop.Raindrop{}
	for _, id := range input.IDs {
		item, err := a.api.Get(ctx, id)
		if err != nil {
			return ai.ClassificationResult{}, err
		}
		items = append(items, item)
	}
	confidence, probability := 0.8, 0.8
	if input.MinConfidence != nil {
		confidence = *input.MinConfidence
	}
	if input.MinProbability != nil {
		probability = *input.MinProbability
	}
	return ai.Classify(ctx, provider, model, items, input.Categories, confidence, probability)
}

func (e *toolEngine) classify(ctx context.Context, args json.RawMessage) *mcp.CallToolResult {
	var input classifyInput
	if err := json.Unmarshal(args, &input); err != nil {
		return toolError(err)
	}
	result, err := (&app{api: e.api}).classifyItems(ctx, input)
	if err != nil {
		return toolError(err)
	}
	return toolJSON(result)
}

func (a *app) classifyCommand(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("classify", flag.ContinueOnError)
	idsText := fs.String("ids", "", "1..25 bookmark IDs")
	taxonomy := fs.String("taxonomy", "", "JSON object mapping category keys to descriptions")
	confidence := fs.Float64("min-confidence", 0.8, "minimum confidence")
	probability := fs.Float64("min-probability", 0.8, "minimum selected-category probability")
	jsonOut := fs.Bool("json", false, "print structured results")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *taxonomy == "" {
		return fmt.Errorf("usage: rdrop classify --ids 1,2 --taxonomy taxonomy.json [--json]")
	}
	ids, err := cli.Int64CSV(*idsText)
	if err != nil {
		return err
	}
	file, err := os.Open(*taxonomy)
	if err != nil {
		return err
	}
	defer file.Close()
	var categories map[string]string
	dec := json.NewDecoder(io.LimitReader(file, 1<<20))
	if err := dec.Decode(&categories); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("taxonomy must contain one JSON object")
	}
	result, err := a.classifyItems(ctx, classifyInput{IDs: ids, Categories: categories, MinConfidence: confidence, MinProbability: probability})
	if err != nil {
		return err
	}
	if *jsonOut {
		return a.print(result, nil)
	}
	for _, item := range result.Results {
		fmt.Fprintf(a.out, "%d\t%s\t%.4f\t%.4f\t%t\n", item.ID, cli.SafeCell(item.Category), item.Confidence, item.Probability, item.NeedsReview)
	}
	return nil
}

func decisionProvider() (ai.Poster, string, error) {
	provider := os.Getenv("RDROP_DECISION_PROVIDER")
	if provider == "" {
		provider = "systemone"
	}
	model := os.Getenv("RDROP_DECISION_MODEL")
	if model == "" {
		model = os.Getenv("TYPESAFE_MODEL")
	}
	if provider == "command" {
		path := os.Getenv("RDROP_DECISION_COMMAND")
		if path == "" {
			return nil, "", fmt.Errorf("set RDROP_DECISION_COMMAND to a trusted executable")
		}
		args := []string{}
		if raw := os.Getenv("RDROP_DECISION_ARGS"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &args); err != nil {
				return nil, "", fmt.Errorf("RDROP_DECISION_ARGS must be a JSON string array")
			}
		}
		return ai.CommandProvider{Path: path, Args: args}, model, nil
	}
	if provider == "cloudflare" {
		base := os.Getenv("RDROP_DECISION_BASE_URL")
		if base == "" {
			base = "https://api.cloudflare.com/client/v4"
		}
		key := os.Getenv("RDROP_DECISION_API_KEY")
		if key == "" {
			key = os.Getenv("CLOUDFLARE_API_TOKEN")
		}
		if key == "" {
			return nil, "", fmt.Errorf("set CLOUDFLARE_API_TOKEN or RDROP_DECISION_API_KEY for Workers AI")
		}
		if os.Getenv("RDROP_DECISION_MODEL") == "" {
			model = "clef-flash"
		}
		client, err := ai.NewCloudflareProvider(base, key, os.Getenv("CLOUDFLARE_ACCOUNT_ID"), model)
		return client, model, err
	}
	if provider != "systemone" {
		return nil, "", fmt.Errorf("RDROP_DECISION_PROVIDER must be systemone, cloudflare or command")
	}
	base := os.Getenv("RDROP_DECISION_BASE_URL")
	if base == "" {
		base = os.Getenv("TYPESAFE_BASE_URL")
	}
	if base == "" {
		base = "https://api.typesafe.ai/v1"
	}
	key := os.Getenv("RDROP_DECISION_API_KEY")
	if key == "" {
		key = os.Getenv("TYPESAFE_API_KEY")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return nil, "", err
	}
	local := parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1"
	if key == "" && !local {
		return nil, "", fmt.Errorf("set RDROP_DECISION_API_KEY (or TYPESAFE_API_KEY); loopback providers may run without a key")
	}
	if model == "" && parsed.Hostname() == "api.typesafe.ai" {
		model = "jev-latest"
	}
	client, err := ai.NewClient(base, key)
	return client, model, err
}
