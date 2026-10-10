package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/KingPsychopath/raindrop-cli/internal/cli"
)

func (a *app) sharing(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("sharing subcommand required")
	}
	switch args[0] {
	case "list":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop sharing list <collection-id>")
		}
		id, err := cli.Int64(args[1], "collection-id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.Sharing(ctx, id))
	case "invite":
		fs := flag.NewFlagSet("sharing invite", flag.ContinueOnError)
		emails := fs.String("emails", "", "comma-separated emails")
		role := fs.String("role", "viewer", "viewer or member")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 || *emails == "" {
			return fmt.Errorf("usage: rdrop sharing invite <collection-id> --emails a@b.com,c@d.com [--role viewer|member]")
		}
		id, err := cli.Int64(fs.Arg(0), "collection-id")
		if err != nil {
			return err
		}
		if err := validateRole(*role); err != nil {
			return err
		}
		return a.printRaw(a.api.Share(ctx, id, cli.CSV(*emails), *role))
	case "role":
		fs := flag.NewFlagSet("sharing role", flag.ContinueOnError)
		role := fs.String("role", "", "viewer or member")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 2 || *role == "" {
			return fmt.Errorf("usage: rdrop sharing role <collection-id> <user-id> --role viewer|member")
		}
		collectionID, err := cli.Int64(fs.Arg(0), "collection-id")
		if err != nil {
			return err
		}
		userID, err := cli.Int64(fs.Arg(1), "user-id")
		if err != nil {
			return err
		}
		if err := validateRole(*role); err != nil {
			return err
		}
		return a.printRaw(a.api.UpdateCollaborator(ctx, collectionID, userID, *role))
	case "remove":
		if len(args) != 3 {
			return fmt.Errorf("usage: rdrop sharing remove <collection-id> <user-id>")
		}
		collectionID, err := cli.Int64(args[1], "collection-id")
		if err != nil {
			return err
		}
		userID, err := cli.Int64(args[2], "user-id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.RemoveCollaborator(ctx, collectionID, userID))
	case "unshare":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop sharing unshare <collection-id>")
		}
		id, err := cli.Int64(args[1], "collection-id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.Unshare(ctx, id))
	case "join":
		fs := flag.NewFlagSet("sharing join", flag.ContinueOnError)
		token := fs.String("token", "", "invitation token")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 || *token == "" {
			return fmt.Errorf("usage: rdrop sharing join <collection-id> --token token")
		}
		id, err := cli.Int64(fs.Arg(0), "collection-id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.JoinCollection(ctx, id, *token))
	default:
		return fmt.Errorf("unknown sharing subcommand %q", args[0])
	}
}
