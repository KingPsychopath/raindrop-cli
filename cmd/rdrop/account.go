package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/machineyearningengineer/raindrop-cli/internal/cli"
)

func (a *app) doctor(ctx context.Context) error {
	raw, err := a.api.User(ctx)
	if err != nil {
		return err
	}
	var payload struct {
		User struct {
			ID       int64  `json:"_id"`
			FullName string `json:"fullName"`
			Email    string `json:"email"`
			Pro      bool   `json:"pro"`
		} `json:"user"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "ok authenticated as %s <%s> id:%d pro:%t\n", payload.User.FullName, payload.User.Email, payload.User.ID, payload.User.Pro)
	return nil
}

func (a *app) user(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("user subcommand required")
	}
	switch args[0] {
	case "update":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop user update '<json>'")
		}
		body, err := cli.JSONObject(args[1])
		if err != nil {
			return err
		}
		return a.printRaw(a.api.UpdateUser(ctx, body))
	default:
		return fmt.Errorf("unknown user subcommand %q", args[0])
	}
}
