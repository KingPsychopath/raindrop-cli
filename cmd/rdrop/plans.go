package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/machineyearningengineer/raindrop-cli/internal/cli"
	"github.com/machineyearningengineer/raindrop-cli/internal/organize"
)

func (a *app) printBatchPlan(ctx context.Context, action string, collection int64, ids []int64, tags []string, to int64) error {
	plan, err := organize.NewPlan(ctx, a.api, action, collection, ids, tags, to)
	return a.print(plan, err)
}

func (a *app) cleanupPlan(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("cleanup plan", flag.ContinueOnError)
	action := fs.String("action", "", "tag, move, or delete")
	collection := fs.Int64("collection", 0, "source collection")
	idsText := fs.String("ids", "", "explicit bookmark IDs")
	tagsText := fs.String("tags", "", "tags to append")
	to := fs.Int64("to", 0, "destination collection")
	fs.Bool("json", true, "plans always print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := noArguments("cleanup plan", fs.Args()); err != nil {
		return err
	}
	ids, err := cli.Int64CSV(*idsText)
	if err != nil {
		return err
	}
	return a.printBatchPlan(ctx, *action, *collection, ids, cli.CSV(*tagsText), *to)
}

func (a *app) cleanupApply(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("cleanup apply", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "apply the reviewed plan")
	dryRun := fs.Bool("dry-run", false, "validate the plan without applying it")
	fs.Bool("json", true, "results always print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 1 || *yes == *dryRun {
		return fmt.Errorf("usage: rdrop cleanup apply <plan.json> --dry-run|--yes")
	}
	var plan organize.Plan
	if err := readJSONFile(fs.Arg(0), &plan); err != nil {
		return fmt.Errorf("decode cleanup plan: %w", err)
	}
	result, applyErr := organize.Apply(ctx, a.api, plan, *yes)
	if err := cli.PrintJSON(a.out, result); err != nil {
		return err
	}
	return applyErr
}
