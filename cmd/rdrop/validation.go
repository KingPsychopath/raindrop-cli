package main

import (
	"flag"
	"fmt"
)

func noArguments(name string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("%s does not accept positional arguments", name)
	}
	return nil
}

func jsonOnly(name string, args []string) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	return noArguments(name, fs.Args())
}

func validatePagination(page, perPage int) error {
	if page < 0 || perPage < 1 || perPage > 50 {
		return fmt.Errorf("--page must be nonnegative and --per-page must be 1..50")
	}
	return nil
}

func validateSource(collection int64) error {
	if collection != -1 && collection != -99 && collection <= 0 {
		return fmt.Errorf("bulk operations require --collection with a positive ID, -1 (Unsorted), or -99 (Trash); the Raindrop API does not support bulk writes to collection 0")
	}
	return nil
}

func validateIDs(ids []int64) error {
	if len(ids) > 100 {
		return fmt.Errorf("select at most 100 explicit IDs per operation")
	}
	seen := map[int64]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return fmt.Errorf("bookmark IDs must be positive and unique (invalid ID %d)", id)
		}
		seen[id] = true
	}
	return nil
}

func validateRole(role string) error {
	if role != "viewer" && role != "member" {
		return fmt.Errorf("--role must be viewer or member")
	}
	return nil
}
