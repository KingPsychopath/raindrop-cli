package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/KingPsychopath/raindrop-cli/internal/cli"
	"github.com/KingPsychopath/raindrop-cli/internal/organize"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func readJSONFile(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, (2<<20)+1))
	if err != nil {
		return err
	}
	if len(raw) > 2<<20 {
		return fmt.Errorf("JSON file exceeds 2 MiB")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON object")
	}
	return nil
}

type contextInput struct {
	IDs        []int64         `json:"ids,omitempty"`
	Collection int64           `json:"collection,omitempty"`
	Search     string          `json:"search,omitempty"`
	Page       int             `json:"page,omitempty"`
	PerPage    int             `json:"perPage,omitempty"`
	Policy     organize.Policy `json:"policy,omitempty"`
}

func (a *app) libraryContext(ctx context.Context, in contextInput) (organize.LibraryContext, error) {
	if in.PerPage == 0 {
		in.PerPage = 25
	}
	if err := validatePagination(in.Page, in.PerPage); err != nil {
		return organize.LibraryContext{}, err
	}
	if in.Collection < 0 && in.Collection != -1 && in.Collection != -99 {
		return organize.LibraryContext{}, fmt.Errorf("invalid collection")
	}
	scope := organize.Scope{CollectionID: in.Collection, Search: in.Search, Page: in.Page, PerPage: in.PerPage, ExplicitIDs: len(in.IDs) > 0}
	if scope.ExplicitIDs && (in.Collection != 0 || in.Search != "" || in.Page != 0) {
		return organize.LibraryContext{}, fmt.Errorf("IDs cannot be combined with scope or page filters")
	}
	if !scope.ExplicitIDs {
		items, _, err := a.api.List(ctx, raindrop.ListOptions{CollectionID: in.Collection, Search: in.Search, Page: in.Page, PerPage: in.PerPage})
		if err != nil {
			return organize.LibraryContext{}, err
		}
		for _, item := range items {
			in.IDs = append(in.IDs, item.ID)
		}
	}
	return organize.NewContext(ctx, a.api, in.IDs, scope, in.Policy)
}
func (a *app) contextCommand(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("context", flag.ContinueOnError)
	idsText := fs.String("ids", "", "select exact bookmark IDs")
	collection := fs.Int64("collection", 0, "collection scope")
	search := fs.String("search", "", "search scope")
	page := fs.Int("page", 0, "zero-based page")
	perPage := fs.Int("per-page", 25, "1..50 bookmarks")
	policyFile := fs.String("policy", "", "JSON naming rules")
	modelOnly := fs.Bool("model-view", false, "omit private notes and apply snapshots")
	fs.Bool("json", true, "context always prints JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := noArguments("context", fs.Args()); err != nil {
		return err
	}
	ids, err := cli.Int64CSV(*idsText)
	if err != nil {
		return err
	}
	var policy organize.Policy
	if *policyFile != "" {
		if err := readJSONFile(*policyFile, &policy); err != nil {
			return err
		}
	}
	c, err := a.libraryContext(ctx, contextInput{IDs: ids, Collection: *collection, Search: *search, Page: *page, PerPage: *perPage, Policy: policy})
	if err != nil {
		return err
	}
	if *modelOnly {
		c = c.ModelView()
	}
	return a.print(c, nil)
}
func (a *app) organizeCommand(args []string) error {
	if len(args) == 0 || args[0] != "plan" {
		return fmt.Errorf("usage: rdrop organize plan --context context.json --proposal proposal.json")
	}
	fs := flag.NewFlagSet("organize plan", flag.ContinueOnError)
	contextFile := fs.String("context", "", "original library context JSON")
	proposalFile := fs.String("proposal", "", "provider-neutral proposal JSON")
	fs.Bool("json", true, "plans always print JSON")
	if err := parseFlags(fs, args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 || *contextFile == "" || *proposalFile == "" {
		return fmt.Errorf("--context and --proposal are required")
	}
	var c organize.LibraryContext
	var p organize.Proposal
	if err := readJSONFile(*contextFile, &c); err != nil {
		return err
	}
	if err := readJSONFile(*proposalFile, &p); err != nil {
		return err
	}
	result, err := organize.PlanProposal(c, p)
	return a.print(result, err)
}

// Each process retains up to 16 contexts. Context responses omit raw snapshots; saved plans include them.
type contextCache struct {
	sync.Mutex
	entries map[string]organize.LibraryContext
	order   []string
}

func (c *contextCache) put(value organize.LibraryContext) {
	c.Lock()
	defer c.Unlock()
	if c.entries == nil {
		c.entries = map[string]organize.LibraryContext{}
	}
	if _, ok := c.entries[value.Digest]; !ok {
		c.order = append(c.order, value.Digest)
	}
	c.entries[value.Digest] = value
	if len(c.order) > 16 {
		delete(c.entries, c.order[0])
		c.order = c.order[1:]
	}
}
func (c *contextCache) get(digest string) (organize.LibraryContext, bool) {
	c.Lock()
	defer c.Unlock()
	value, ok := c.entries[digest]
	return value, ok
}
func contextTools() []*mcp.Tool {
	return []*mcp.Tool{
		{Name: "raindrop_context", Description: "Read bounded library context: exact existing tags, collection paths and bookmark metadata. Private notes and apply snapshots remain in process. Use this digest for raindrop_organize_plan. Defaults to one page of 25; not a whole-library scan.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "maxItems": 50, "minItems": 1, "uniqueItems": true}, "collection": map[string]any{"type": "integer"}, "search": map[string]any{"type": "string"}, "page": map[string]any{"type": "integer", "minimum": 0}, "perPage": map[string]any{"type": "integer", "minimum": 1, "maximum": 50}}, "additionalProperties": false}, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPointer(false)}},
		{Name: "raindrop_organize_plan", Description: "Validate multi-tag and move suggestions against a retained context. Preserves existing tags, rejects unknown names/destinations/IDs, withholds needsReview entries, returns exact snapshot plan. Never applies changes. Context cache is process-local and bounded.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"contextDigest": map[string]any{"type": "string"}, "suggestions": map[string]any{"type": "array", "maxItems": 50, "items": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "integer", "minimum": 1}, "tags": map[string]any{"type": "array", "items": map[string]any{"type": "string", "minLength": 1}, "uniqueItems": true}, "collectionId": map[string]any{"type": "integer"}, "reason": map[string]any{"type": "string", "minLength": 1}, "needsReview": map[string]any{"type": "boolean"}}, "required": []string{"id", "reason", "needsReview"}, "additionalProperties": false}}}, "required": []string{"contextDigest", "suggestions"}, "additionalProperties": false}, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPointer(false)}},
	}
}
func (e *toolEngine) contextCall(ctx context.Context, name string, args json.RawMessage) *mcp.CallToolResult {
	if name == "raindrop_context" {
		var in contextInput
		if err := json.Unmarshal(args, &in); err != nil {
			return toolError(err)
		}
		c, err := (&app{api: e.api}).libraryContext(ctx, in)
		if err != nil {
			return toolError(err)
		}
		e.contexts.put(c)
		return toolJSON(c.ModelView())
	}
	var p organize.Proposal
	if err := json.Unmarshal(args, &p); err != nil {
		return toolError(err)
	}
	c, ok := e.contexts.get(p.ContextDigest)
	if !ok {
		return toolError(fmt.Errorf("context expired or unavailable; read raindrop_context again"))
	}
	result, err := organize.PlanProposal(c, p)
	if err != nil {
		return toolError(err)
	}
	return toolJSON(result)
}
