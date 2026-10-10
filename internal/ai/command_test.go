package ai

import (
	"context"
	"encoding/json"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
	"os"
	"testing"
	"time"
)

func TestDecisionCommandHelper(t *testing.T) {
	if os.Getenv("RDROP_COMMAND_FIXTURE") != "1" {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
		os.Exit(3)
	}
	if os.Getenv("RAINDROP_TOKEN") != "" || body["questions"] == nil {
		os.Exit(4)
	}
	os.Stdout.Write([]byte(`{"model":"local-fixture","answers":{"42":{"type":"choice","choice":"docs","confidence":0.99,"probabilities":{"docs":0.95,"needs_review":0.05},"abstention":true}}}`))
	os.Exit(0)
}
func TestLocalCommandUsesSameValidation(t *testing.T) {
	t.Setenv("RDROP_COMMAND_FIXTURE", "1")
	t.Setenv("RAINDROP_TOKEN", "secret")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	provider := CommandProvider{Path: executable, Args: []string{"-test.run=^TestDecisionCommandHelper$"}, Timeout: 5 * time.Second}
	result, err := Classify(context.Background(), provider, "local", []raindrop.Raindrop{{ID: 42}}, map[string]string{"docs": "Docs"}, 0.8, 0.8)
	if err != nil || len(result.Results) != 1 || !result.Results[0].NeedsReview {
		t.Fatalf("%+v %v", result, err)
	}
	provider.Args = []string{"-invalid-test-flag"}
	if _, err := Classify(context.Background(), provider, "local", []raindrop.Raindrop{{ID: 42}}, map[string]string{"docs": "Docs"}, 0.8, 0.8); err == nil {
		t.Fatal("command error accepted")
	}
}
