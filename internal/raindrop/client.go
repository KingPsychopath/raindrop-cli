package raindrop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.raindrop.io/rest/v1"

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	data, _, err := c.do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if raw, ok := out.(*json.RawMessage); ok {
		*raw = append((*raw)[:0], data...)
		return nil
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) Bytes(ctx context.Context, method, path string, query url.Values, body any) ([]byte, string, error) {
	return c.do(ctx, method, path, query, body)
}

func (c *Client) Multipart(ctx context.Context, method, path string, query url.Values, fileField, filePath string, fields map[string]string) (json.RawMessage, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create multipart file field: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy multipart file: %w", err)
	}
	for key, value := range fields {
		if value == "" {
			continue
		}
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("write multipart field %s: %w", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	u, err := c.url(path, query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), &body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	c.setCommonHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	data, _, err := c.send(req)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) ([]byte, string, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, "", fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	u, err := c.url(path, query)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, "", fmt.Errorf("build request: %w", err)
	}
	c.setCommonHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.send(req)
}

func (c *Client) url(path string, query url.Values) (*url.URL, error) {
	u, err := url.Parse(c.baseURL + "/" + strings.TrimLeft(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u, nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "raindrop-cli/0.1")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func (c *Client) send(req *http.Request) ([]byte, string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request %s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, "", apiError(resp.StatusCode, data)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

func apiError(status int, data []byte) error {
	var payload struct {
		Error        any    `json:"error"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(data, &payload); err == nil && payload.ErrorMessage != "" {
		return fmt.Errorf("raindrop api status %d: %s", status, payload.ErrorMessage)
	}
	body := strings.TrimSpace(string(data))
	if body == "" {
		body = http.StatusText(status)
	}
	return fmt.Errorf("raindrop api status %d: %s", status, body)
}
