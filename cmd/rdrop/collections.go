package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/KingPsychopath/raindrop-cli/internal/cli"
)

func (a *app) collections(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("collections", flag.ContinueOnError)
	children := fs.Bool("children", false, "include child collections")
	tree := fs.Bool("tree", false, "include all collections with breadcrumb paths")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	a.json = *jsonOut
	if *tree {
		items, err := a.api.CollectionTree(ctx)
		if err != nil {
			return err
		}
		if a.json {
			return a.print(items, nil)
		}
		for _, item := range items {
			fmt.Fprintf(a.out, "%d\t%s\t%d\n", item.ID, cli.SafeCell(item.Path), item.Count)
		}
		return nil
	}
	if err := noArguments("collections", fs.Args()); err != nil {
		return err
	}
	items, err := a.api.Collections(ctx, *children)
	if a.json {
		return a.print(items, err)
	}
	if err != nil {
		return err
	}
	cli.PrintCollections(a.out, items)
	return nil
}

func (a *app) collection(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("collection subcommand required")
	}
	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("collection create", flag.ContinueOnError)
		parent := fs.Int64("parent", 0, "parent collection id")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: rdrop collection create [--parent id] <title>")
		}
		return a.printRaw(a.api.CreateCollection(ctx, fs.Arg(0), *parent))
	case "update":
		if len(args) != 3 {
			return fmt.Errorf("usage: rdrop collection update <id> '<json>'")
		}
		id, err := cli.Int64(args[1], "id")
		if err != nil {
			return err
		}
		body, err := cli.JSONObject(args[2])
		if err != nil {
			return err
		}
		return a.printRaw(a.api.UpdateCollection(ctx, id, body))
	case "delete":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop collection delete <id>")
		}
		id, err := cli.Int64(args[1], "id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.DeleteCollection(ctx, id))
	case "sort":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop collection sort title|-title|-count")
		}
		return a.printRaw(a.api.SortCollections(ctx, args[1]))
	case "expand":
		if len(args) != 2 || (args[1] != "true" && args[1] != "false") {
			return fmt.Errorf("usage: rdrop collection expand true|false")
		}
		return a.printRaw(a.api.ExpandCollections(ctx, args[1] == "true"))
	case "merge":
		fs := flag.NewFlagSet("collection merge", flag.ContinueOnError)
		to := fs.Int64("to", 0, "destination collection id")
		idsFlag := fs.String("ids", "", "comma-separated collection ids")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		ids, err := cli.Int64CSV(*idsFlag)
		if err != nil {
			return err
		}
		if *to == 0 || len(ids) == 0 {
			return fmt.Errorf("usage: rdrop collection merge --to id --ids id,id")
		}
		return a.printRaw(a.api.MergeCollections(ctx, *to, ids))
	case "clean":
		if len(args) != 1 {
			return fmt.Errorf("usage: rdrop collection clean")
		}
		return a.printRaw(a.api.CleanCollections(ctx))
	case "empty-trash":
		fs := flag.NewFlagSet("collection empty-trash", flag.ContinueOnError)
		yes := fs.Bool("yes", false, "confirm emptying Trash")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || !*yes {
			return fmt.Errorf("usage: rdrop collection empty-trash --yes")
		}
		return a.printRaw(a.api.DeleteCollection(ctx, -99))
	default:
		return fmt.Errorf("unknown collection subcommand %q", args[0])
	}
}

func (a *app) covers(ctx context.Context, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: rdrop covers [search-text]")
	}
	search := ""
	if len(args) == 1 {
		search = args[0]
	}
	return a.printRaw(a.api.CollectionCovers(ctx, search))
}
