package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/KingPsychopath/raindrop-cli/internal/cli"
)

func (a *app) tags(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("tags", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	a.json = *jsonOut
	if err := noArguments("tags", fs.Args()); err != nil {
		return err
	}
	items, err := a.api.Tags(ctx, *collection)
	if a.json {
		return a.print(items, err)
	}
	if err != nil {
		return err
	}
	cli.PrintTags(a.out, items)
	return nil
}

func (a *app) tag(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("tag subcommand required")
	}
	fs := flag.NewFlagSet("tag "+args[0], flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	if err := parseFlags(fs, args[1:]); err != nil {
		return err
	}
	switch args[0] {
	case "rename":
		if fs.NArg() != 2 {
			return fmt.Errorf("usage: rdrop tag rename [--collection id] <old> <new>")
		}
		return a.printRaw(a.api.RenameTags(ctx, *collection, []string{fs.Arg(0)}, fs.Arg(1)))
	case "merge":
		if fs.NArg() != 2 {
			return fmt.Errorf("usage: rdrop tag merge [--collection id] <old,old2> <new>")
		}
		return a.printRaw(a.api.RenameTags(ctx, *collection, cli.CSV(fs.Arg(0)), fs.Arg(1)))
	case "delete":
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: rdrop tag delete [--collection id] <tag,tag2>")
		}
		return a.printRaw(a.api.DeleteTags(ctx, *collection, cli.CSV(fs.Arg(0))))
	default:
		return fmt.Errorf("unknown tag subcommand %q", args[0])
	}
}

func (a *app) filters(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("filters", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	search := fs.String("search", "", "search query")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := noArguments("filters", fs.Args()); err != nil {
		return err
	}
	a.json = *jsonOut
	return a.print(a.api.Filters(ctx, *collection, *search))
}

func (a *app) cleanup(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		return a.help([]string{"cleanup"})
	}
	switch args[0] {
	case "report":
		fs := flag.NewFlagSet("cleanup report", flag.ContinueOnError)
		collection := fs.Int64("collection", 0, "collection id")
		search := fs.String("search", "", "search query")
		jsonOut := fs.Bool("json", false, "print JSON")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		filters, err := a.api.Filters(ctx, *collection, *search)
		if err != nil {
			return err
		}
		report := cleanupReport(*collection, *search, filters)
		if *jsonOut {
			return cli.PrintJSON(a.out, report)
		}
		printCleanupReport(a.out, report)
		return nil
	case "plan":
		return a.cleanupPlan(ctx, args[1:])
	case "apply":
		return a.cleanupApply(ctx, args[1:])
	default:
		return fmt.Errorf("unknown cleanup subcommand %q", args[0])
	}
}
