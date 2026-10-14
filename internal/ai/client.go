// Package ai contains bounded HTTP clients for optional model integrations.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/machineyearningengineer/raindrop-cli/internal/config"
)

type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewClient(baseURL, key string) (*Client, error) {
	if err := config.ValidateURL(baseURL); err != nil {
		return nil, err
	}
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Key: key, HTTP: &http.Client{Timeout: 120 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return fmt.Errorf("model API redirects are not supported")
	}}}, nil
}

func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/"+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("model request failed: %w", err)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
	if err != nil {
		return err
	}
	if len(response) > 8<<20 {
		return fmt.Errorf("model response exceeds 8 MiB limit")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Avoid echoing upstream bodies: they can contain credentials or library data.
		return fmt.Errorf("model API HTTP %d (retry-after: %s)", resp.StatusCode, resp.Header.Get("Retry-After"))
	}
	if err := json.Unmarshal(response, out); err != nil {
		return fmt.Errorf("decode model response: %w", err)
	}
	return nil
}
