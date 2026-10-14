package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/machineyearningengineer/raindrop-cli/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const officialMCPEndpoint = "https://api.raindrop.io/rest/v2/ai/mcp"

// Only documented reads are imported. Upstream annotations cannot grant writes.
var officialReads = map[string]bool{
	"find_bookmarks": true, "fetch_bookmark_content": true,
	"find_misplaced_bookmarks": true, "find_mistagged_bookmarks": true,
	"find_collections": true, "find_tags": true, "find_highlights": true,
	"fetch_current_user": true, "fetch_popular_keywords": true,
}

type officialBridge struct {
	session *mcp.ClientSession
	tools   []*mcp.Tool
	names   map[string]string
}

type officialAuthTransport struct {
	origin *url.URL
	token  string
	next   http.RoundTripper
}

func (t officialAuthTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != t.origin.Scheme || request.URL.Host != t.origin.Host {
		return nil, fmt.Errorf("refusing official MCP credentials outside its origin")
	}
	clone := request.Clone(request.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	response, err := t.next.RoundTrip(clone)
	if err == nil {
		response.Body = &boundedOfficialBody{ReadCloser: response.Body, reader: io.LimitReader(response.Body, 8<<20)}
	}
	return response, err
}

type boundedOfficialBody struct {
	io.ReadCloser
	reader io.Reader
}

func (b *boundedOfficialBody) Read(p []byte) (int, error) { return b.reader.Read(p) }

func connectOfficial(ctx context.Context, endpoint, token string) (*officialBridge, error) {
	if err := config.ValidateURL(endpoint); err != nil {
		return nil, err
	}
	origin, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "raindrop-cli-official-reads", Version: version}, &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}})
	httpClient := &http.Client{Timeout: 120 * time.Second, Transport: officialAuthTransport{origin, token, http.DefaultTransport}, CheckRedirect: func(*http.Request, []*http.Request) error {
		return fmt.Errorf("official MCP redirects are not supported")
	}}
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, HTTPClient: httpClient, DisableStandaloneSSE: true, MaxRetries: -1, MaxEventSize: 8 << 20}, nil)
	if err != nil {
		return nil, fmt.Errorf("official MCP initialization failed: %w", err)
	}
	bridge := &officialBridge{session: session, tools: []*mcp.Tool{}, names: map[string]string{}}
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			session.Close()
			return nil, fmt.Errorf("official MCP discovery failed: %w", err)
		}
		if !officialReads[tool.Name] {
			continue
		}
		copy := *tool
		copy.Name = "raindrop_official_" + tool.Name
		copy.Description = "Official Raindrop hosted read; account plan/service availability applies. " + tool.Description
		copy.Annotations = &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: boolPointer(false)}
		bridge.tools = append(bridge.tools, &copy)
		bridge.names[copy.Name] = tool.Name
	}
	if len(bridge.tools) == 0 {
		session.Close()
		return nil, fmt.Errorf("official MCP exposed no supported read tools")
	}
	return bridge, nil
}

func (e *toolEngine) enableOfficial(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	// Deliberately fixed: configuration for another MCP server must not receive
	// this Raindrop credential. Loopback fixtures use connectOfficial directly.
	e.official, err = connectOfficial(ctx, officialMCPEndpoint, cfg.Token)
	return err
}

func (b *officialBridge) call(ctx context.Context, name string, input json.RawMessage) *mcp.CallToolResult {
	original, ok := b.names[name]
	if !ok || !officialReads[original] {
		return toolError(fmt.Errorf("official mutation or unknown tool is unavailable"))
	}
	result, err := b.session.CallTool(ctx, &mcp.CallToolParams{Name: original, Arguments: input})
	if err != nil {
		return toolError(fmt.Errorf("official MCP read failed: %w", err))
	}
	return result
}

func isOfficialTool(name string) bool { return strings.HasPrefix(name, "raindrop_official_") }
