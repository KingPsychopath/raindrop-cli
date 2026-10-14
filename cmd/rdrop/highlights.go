package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/machineyearningengineer/raindrop-cli/internal/cli"
)

func (a *app) highlights(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("highlights", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	page := fs.Int("page", 0, "page number")
	perPage := fs.Int("per-page", 25, "results per page")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := noArguments("highlights", fs.Args()); err != nil {
		return err
	}
	if err := validatePagination(*page, *perPage); err != nil {
		return err
	}
	return a.printRaw(a.api.Highlights(ctx, *collection, *page, *perPage))
}

func (a *app) highlight(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("highlight subcommand required")
	}
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("highlight add", flag.ContinueOnError)
		text := fs.String("text", "", "highlight text")
		color := fs.String("color", "", "highlight color")
		note := fs.String("note", "", "highlight note")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 || *text == "" {
			return fmt.Errorf("usage: rdrop highlight add <raindrop-id> --text text [--color yellow] [--note text]")
		}
		id, err := cli.Int64(fs.Arg(0), "raindrop-id")
		if err != nil {
			return err
		}
		highlight := map[string]any{"text": *text}
		if *color != "" {
			highlight["color"] = *color
		}
		if *note != "" {
			highlight["note"] = *note
		}
		return a.print(a.api.Update(ctx, id, map[string]any{"highlights": []any{highlight}}))
	case "update":
		fs := flag.NewFlagSet("highlight update", flag.ContinueOnError)
		text := fs.String("text", "", "highlight text")
		color := fs.String("color", "", "highlight color")
		note := fs.String("note", "", "highlight note")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 2 {
			return fmt.Errorf("usage: rdrop highlight update <raindrop-id> <highlight-id> [--text text] [--color yellow] [--note text]")
		}
		id, err := cli.Int64(fs.Arg(0), "raindrop-id")
		if err != nil {
			return err
		}
		highlight := map[string]any{"_id": fs.Arg(1)}
		if *text != "" {
			highlight["text"] = *text
		}
		if *color != "" {
			highlight["color"] = *color
		}
		if *note != "" {
			highlight["note"] = *note
		}
		return a.print(a.api.Update(ctx, id, map[string]any{"highlights": []any{highlight}}))
	case "delete":
		if len(args) != 3 {
			return fmt.Errorf("usage: rdrop highlight delete <raindrop-id> <highlight-id>")
		}
		id, err := cli.Int64(args[1], "raindrop-id")
		if err != nil {
			return err
		}
		return a.print(a.api.Update(ctx, id, map[string]any{
			"highlights": []any{map[string]any{"_id": args[2], "text": ""}},
		}))
	default:
		return fmt.Errorf("unknown highlight subcommand %q", args[0])
	}
}
