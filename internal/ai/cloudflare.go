package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

// CloudflareProvider adapts the Workers AI envelope to System One decisions.
type CloudflareProvider struct {
	client *Client
	path   string
}

func NewCloudflareProvider(base, key, account, model string) (*CloudflareProvider, error) {
	if !regexp.MustCompile(`^[a-fA-F0-9]{32}$`).MatchString(account) {
		return nil, fmt.Errorf("CLOUDFLARE_ACCOUNT_ID must be a 32-character hexadecimal account ID")
	}
	if model != "clef" && model != "clef-flash" && model != "clef-omni" {
		return nil, fmt.Errorf("Cloudflare decision model must be clef, clef-flash or clef-omni")
	}
	client, err := NewClient(base, key)
	if err != nil {
		return nil, err
	}
	return &CloudflareProvider{client: client, path: "accounts/" + account + "/ai/run/@cf/cloudflare/" + model}, nil
}

func (p *CloudflareProvider) Post(ctx context.Context, method string, body, out any) error {
	if method != "systemone" {
		return fmt.Errorf("Cloudflare decision adapter only supports systemone")
	}
	var envelope struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}
	if err := p.client.Post(ctx, p.path, body, &envelope); err != nil {
		return err
	}
	if !envelope.Success || len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return fmt.Errorf("Cloudflare did not return a successful decision result")
	}
	return json.Unmarshal(envelope.Result, out)
}
