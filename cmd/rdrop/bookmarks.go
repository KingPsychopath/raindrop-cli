package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/machineyearningengineer/raindrop-cli/internal/cli"
	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
	"strings"
)

func (a *app) list(ctx context.Context, command string, args []string) error {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	search := fs.String("search", "", "search query")
	page := fs.Int("page", 0, "page number")
	perPage := fs.Int("per-page", 50, "results per page")
	sort := fs.String("sort", "-created", "sort")
	nested := fs.Bool("nested", false, "include nested collections")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := validatePagination(*page, *perPage); err != nil {
		return err
	}
	if command != "search" {
		if err := noArguments(command, fs.Args()); err != nil {
			return err
		}
	}
	if command == "search" {
		if fs.NArg() < 1 {
			return fmt.Errorf("usage: rdrop search <query> [flags]")
		}
		*search = strings.Join(fs.Args(), " ")
	}
	if command == "untagged" {
		*search = appendSearch(*search, "notag:true")
	}
	if command == "duplicates" {
		*search = appendSearch(*search, "duplicates:true")
	}
	if command == "broken" {
		*search = appendSearch(*search, "broken:true")
	}
	a.json = *jsonOut
	items, count, err := a.api.List(ctx, raindrop.ListOptions{
		CollectionID: *collection,
		Search:       *search,
		Page:         *page,
		PerPage:      *perPage,
		Sort:         *sort,
		Nested:       *nested,
	})
	if a.json {
		return a.print(map[string]any{"count": count, "items": items}, err)
	}
	if err != nil {
		return err
	}
	cli.PrintRaindrops(a.out, items)
	return nil
}

func appendSearch(existing, extra string) string {
	if existing == "" {
		return extra
	}
	return existing + " " + extra
}

func (a *app) get(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: rdrop get [--json] <id>")
	}
	a.json = *jsonOut
	id, err := cli.Int64(fs.Arg(0), "id")
	if err != nil {
		return err
	}
	item, err := a.api.Get(ctx, id)
	if a.json {
		return a.print(item, err)
	}
	if err != nil {
		return err
	}
	cli.PrintRaindrops(a.out, []raindrop.Raindrop{item})
	return nil
}

func (a *app) add(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	title := fs.String("title", "", "title")
	tags := fs.String("tags", "", "comma-separated tags")
	collection := fs.Int64("collection", 0, "collection id")
	note := fs.String("note", "", "note")
	excerpt := fs.String("excerpt", "", "excerpt")
	parse := fs.Bool("parse", false, "ask Raindrop to parse metadata")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: rdrop add <url> [flags]")
	}
	a.json = *jsonOut
	body := map[string]any{"link": fs.Arg(0)}
	if *title != "" {
		body["title"] = *title
	}
	if *tags != "" {
		body["tags"] = cli.CSV(*tags)
	}
	if *collection != 0 {
		body["collection"] = raindrop.Ref{ID: *collection}
	}
	if *note != "" {
		body["note"] = *note
	}
	if *excerpt != "" {
		body["excerpt"] = *excerpt
	}
	if *parse {
		body["pleaseParse"] = map[string]any{}
	}
	return a.print(a.api.Create(ctx, body))
}

func (a *app) update(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: rdrop update <id> '<json>'")
	}
	id, err := cli.Int64(args[0], "id")
	if err != nil {
		return err
	}
	body, err := cli.JSONObject(args[1])
	if err != nil {
		return err
	}
	return a.print(a.api.Update(ctx, id, body))
}

func (a *app) delete(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: rdrop delete <id>")
	}
	id, err := cli.Int64(args[0], "id")
	if err != nil {
		return err
	}
	return a.printRaw(a.api.Delete(ctx, id))
}

func (a *app) suggest(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("suggest", flag.ContinueOnError)
	fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	args = fs.Args()
	if len(args) != 1 {
		return fmt.Errorf("usage: rdrop suggest <url>")
	}
	return a.printRaw(a.api.Suggest(ctx, args[0]))
}
