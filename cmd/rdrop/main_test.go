package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

func TestParseFlagsAllowsFlagsAfterPositionals(t *testing.T) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	title := fs.String("title", "", "title")
	tags := fs.String("tags", "", "tags")
	parse := fs.Bool("parse", false, "parse")

	err := parseFlags(fs, []string{"https://example.com", "--title", "Example", "--tags=docs,go", "--parse"})
	if err != nil {
		t.Fatal(err)
	}
	if fs.NArg() != 1 || fs.Arg(0) != "https://example.com" {
		t.Fatalf("args = %#v", fs.Args())
	}
	if *title != "Example" || *tags != "docs,go" || !*parse {
		t.Fatalf("title=%q tags=%q parse=%t", *title, *tags, *parse)
	}
}

func TestCompletionDoesNotRequireToken(t *testing.T) {
	t.Setenv("RAINDROP_TOKEN", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := run(context.Background(), []string{"completion", "bash"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseFlagsKeepsFlagValueThatLooksLikeFlag(t *testing.T) {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	sort := fs.String("sort", "", "sort")

	err := parseFlags(fs, []string{"--sort", "-created"})
	if err != nil {
		t.Fatal(err)
	}
	if *sort != "-created" {
		t.Fatalf("sort = %q", *sort)
	}
}

func TestHelpTopicsCoverTopLevelCommands(t *testing.T) {
	commands := []string{
		"doctor", "me", "user", "stats", "collections", "collection", "covers", "cover",
		"export", "import", "upload", "tags", "tag", "filters", "cleanup", "list",
		"search", "get", "add", "update", "delete", "batch", "highlights", "highlight",
		"cache", "sharing", "backups", "backup", "suggest", "completion", "raw",
	}
	for _, command := range commands {
		if helpTopics[command] == "" {
			t.Fatalf("missing help topic for %s", command)
		}
	}
}

func TestTagNormalizationPlans(t *testing.T) {
	plans := tagNormalizationPlans(123, []raindrop.Tag{
		{ID: "code style", Count: 6},
		{ID: "TIL", Count: 4},
	})
	if len(plans) != 2 {
		t.Fatalf("len(plans) = %d", len(plans))
	}
	if !strings.Contains(plans[0].Command, "'code style'") {
		t.Fatalf("command did not quote spaced tag: %q", plans[0].Command)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"dev":        "dev",
		"code style": "'code style'",
		"bob's":      "'bob'\"'\"'s'",
	}
	for input, want := range cases {
		if got := shellQuote(input); got != want {
			t.Fatalf("shellQuote(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBatchDeleteRequiresDryRunOrYes(t *testing.T) {
	err := (&app{out: os.Stdout, err: os.Stderr}).batch(context.Background(), []string{"delete", "--ids", "1,2"})
	if err == nil || !strings.Contains(err.Error(), "--dry-run or --yes") {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchTagAndMoveRequireIDs(t *testing.T) {
	err := (&app{out: os.Stdout, err: os.Stderr}).batch(context.Background(), []string{"tag", "--tags", "go"})
	if err == nil || !strings.Contains(err.Error(), "batch tag") {
		t.Fatalf("tag err = %v", err)
	}
	err = (&app{out: os.Stdout, err: os.Stderr}).batch(context.Background(), []string{"move", "--to", "123"})
	if err == nil || !strings.Contains(err.Error(), "batch move") {
		t.Fatalf("move err = %v", err)
	}
}

func TestBatchDeleteApplyCommand(t *testing.T) {
	got := batchDeleteApplyCommand(123, "notag:true", []int64{1, 2})
	want := "rdrop batch delete --ids 1,2 --search notag:true --collection 123 --yes"
	if got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}
