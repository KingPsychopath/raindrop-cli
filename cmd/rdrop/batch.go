package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/KingPsychopath/raindrop-cli/internal/cli"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
	"strings"
)

func (a *app) batch(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("batch subcommand required")
	}
	fs := flag.NewFlagSet("batch "+args[0], flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "source collection id")
	idsFlag := fs.String("ids", "", "comma-separated bookmark ids")
	tagsFlag := fs.String("tags", "", "comma-separated tags")
	to := fs.Int64("to", 0, "destination collection id")
	search := fs.String("search", "", "search query")
	dryRun := fs.Bool("dry-run", false, "preview the bulk operation without mutating")
	yes := fs.Bool("yes", false, "confirm a destructive bulk delete")
	jsonOut := fs.Bool("json", false, "print JSON for dry-run output")
	if err := parseFlags(fs, args[1:]); err != nil {
		return err
	}
	if err := noArguments("batch", fs.Args()); err != nil {
		return err
	}
	if *dryRun && *yes {
		return fmt.Errorf("choose --dry-run or --yes, not both")
	}
	if *idsFlag != "" && *search != "" {
		return fmt.Errorf("choose --ids or --search, not both")
	}
	ids, err := cli.Int64CSV(*idsFlag)
	if err != nil {
		return err
	}
	if err := validateIDs(ids); err != nil {
		return err
	}
	switch args[0] {
	case "tag":
		if len(ids) == 0 || len(cli.CSV(*tagsFlag)) == 0 {
			return fmt.Errorf("usage: rdrop batch tag --ids 1,2 --tags go,docs [--collection id]")
		}
		if err := validateSource(*collection); err != nil {
			return err
		}
		if *dryRun {
			return a.printBatchPlan(ctx, "tag", *collection, ids, cli.CSV(*tagsFlag), 0)
		}
		body := map[string]any{"tags": cli.CSV(*tagsFlag)}
		body["ids"] = ids
		return a.printRaw(a.api.BatchUpdate(ctx, *collection, body))
	case "move":
		if len(ids) == 0 || *to == 0 {
			return fmt.Errorf("usage: rdrop batch move --ids 1,2 --to collectionID [--collection id]")
		}
		if err := validateSource(*collection); err != nil {
			return err
		}
		if *to != -1 && *to <= 0 {
			return fmt.Errorf("--to must be a positive collection ID or -1")
		}
		if *dryRun {
			return a.printBatchPlan(ctx, "move", *collection, ids, nil, *to)
		}
		body := map[string]any{"collection": raindrop.Ref{ID: *to}, "ids": ids}
		return a.printRaw(a.api.BatchUpdate(ctx, *collection, body))
	case "delete":
		if len(ids) == 0 && *search == "" {
			return fmt.Errorf("usage: rdrop batch delete --ids 1,2|--search query [--collection id] --dry-run|--yes")
		}
		if *dryRun {
			if err := validateSource(*collection); err != nil {
				return err
			}
			return a.printBatchDeletePlan(ctx, *collection, *search, ids, *jsonOut)
		}
		if !*yes {
			return fmt.Errorf("refusing bulk delete without --dry-run or --yes")
		}
		if err := validateSource(*collection); err != nil {
			return err
		}
		return a.printRaw(a.api.BatchDelete(ctx, *collection, *search, ids))
	default:
		return fmt.Errorf("unknown batch subcommand %q", args[0])
	}
}

func (a *app) printBatchDeletePlan(ctx context.Context, collectionID int64, search string, ids []int64, jsonOut bool) error {
	if len(ids) > 0 {
		return a.printBatchPlan(ctx, "delete", collectionID, ids, nil, 0)
	}
	plan := map[string]any{
		"action":       "batch delete",
		"collectionId": collectionID,
		"ids":          ids,
		"search":       search,
		"applyCommand": batchDeleteApplyCommand(collectionID, search, ids),
	}
	if search != "" {
		_, count, err := a.api.List(ctx, raindrop.ListOptions{
			CollectionID: collectionID,
			Search:       search,
			PerPage:      1,
		})
		if err != nil {
			return err
		}
		plan["matchedCount"] = count
	} else {
		plan["matchedCount"] = len(ids)
	}
	if jsonOut {
		return cli.PrintJSON(a.out, plan)
	}
	fmt.Fprintf(a.out, "bulk delete plan\n")
	fmt.Fprintf(a.out, "collection: %d\n", collectionID)
	if search != "" {
		fmt.Fprintf(a.out, "search: %s\n", search)
	}
	if len(ids) > 0 {
		fmt.Fprintf(a.out, "ids: %s\n", int64CSV(ids))
	}
	fmt.Fprintf(a.out, "matched: %v\n", plan["matchedCount"])
	fmt.Fprintf(a.out, "apply: %s\n", plan["applyCommand"])
	return nil
}

func batchDeleteApplyCommand(collectionID int64, search string, ids []int64) string {
	parts := []string{"rdrop", "batch", "delete"}
	if len(ids) > 0 {
		parts = append(parts, "--ids", shellQuote(int64CSV(ids)))
	}
	if search != "" {
		parts = append(parts, "--search", shellQuote(search))
	}
	if collectionID != 0 {
		parts = append(parts, "--collection", fmt.Sprint(collectionID))
	}
	parts = append(parts, "--yes")
	return strings.Join(parts, " ")
}

func int64CSV(values []int64) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprint(value))
	}
	return strings.Join(parts, ",")
}
