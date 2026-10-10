package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Poster interface {
	Post(context.Context, string, any, any) error
}

// CommandProvider executes a user-configured executable directly. No shell is
// involved and model/tool inputs never choose an executable or its arguments.
type CommandProvider struct {
	Path    string
	Args    []string
	Timeout time.Duration
}
type boundedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, fmt.Errorf("provider output exceeds limit")
	}
	return b.Buffer.Write(p)
}
func (c CommandProvider) Post(ctx context.Context, _ string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	command := exec.CommandContext(ctx, c.Path, c.Args...)
	command.Stdin = bytes.NewReader(data)
	stdout := &boundedBuffer{limit: 8 << 20}
	stderr := &boundedBuffer{limit: 64 << 10}
	command.Stdout = stdout
	command.Stderr = stderr
	// A local decision process needs metadata, not the Raindrop account credential.
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if name != "RAINDROP_TOKEN" && name != "TYPESAFE_API_KEY" && name != "OPENAI_API_KEY" && name != "RDROP_MODEL_API_KEY" && name != "RDROP_DECISION_API_KEY" && name != "CLOUDFLARE_API_TOKEN" {
			command.Env = append(command.Env, entry)
		}
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("decision command failed: %w", err)
	}
	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return fmt.Errorf("decode decision command output: %w", err)
	}
	return nil
}
