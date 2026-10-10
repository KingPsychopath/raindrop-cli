package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KingPsychopath/raindrop-cli/internal/cli"
	"github.com/KingPsychopath/raindrop-cli/internal/config"
)

func (a *app) config(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		return a.help([]string{"config"})
	}
	if len(args) != 1 || (args[0] != "paths" && args[0] != "show") {
		return fmt.Errorf("usage: rdrop config paths|show")
	}
	if args[0] == "paths" {
		return cli.PrintJSON(a.out, map[string]any{"envFiles": config.EnvFiles(), "tokenFileVariable": "RAINDROP_TOKEN_FILE"})
	}
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	return cli.PrintJSON(a.out, cfg.Status())
}

func (a *app) connect(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		return a.help([]string{"connect"})
	}
	fs := flag.NewFlagSet("connect", flag.ContinueOnError)
	executable, _ := os.Executable()
	command := fs.String("command", executable, "path to installed rdrop")
	cwd := fs.String("cwd", "", "ACP session working directory")
	endpoint := fs.String("url", "", "HTTP MCP endpoint")
	write := fs.Bool("allow-write", false, "include write tools in the local server")
	if err := parseFlags(fs, args[1:]); err != nil {
		return err
	}
	if err := noArguments("connect", fs.Args()); err != nil {
		return err
	}
	serverArgs := []string{"mcp", "serve"}
	if *write {
		serverArgs = append(serverArgs, "--allow-write")
	}
	switch args[0] {
	case "mcp":
		return cli.PrintJSON(a.out, map[string]any{"mcpServers": map[string]any{"raindrop-cli": map[string]any{"command": *command, "args": serverArgs}}})
	case "acp":
		if *cwd == "" {
			var err error
			*cwd, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		absolute, err := filepath.Abs(*cwd)
		if err != nil {
			return err
		}
		return cli.PrintJSON(a.out, map[string]any{"cwd": absolute, "mcpServers": []any{map[string]any{"name": "raindrop-cli", "command": *command, "args": serverArgs, "env": []any{}}}})
	case "raindrop", "executor":
		if *write {
			return fmt.Errorf("--allow-write controls only local rdrop MCP; configure remote permissions in the remote service")
		}
		if *endpoint == "" {
			if args[0] == "executor" {
				*endpoint = "http://127.0.0.1:4312/mcp"
			} else {
				*endpoint = os.Getenv("RAINDROP_MCP_ENDPOINT")
				if *endpoint == "" {
					*endpoint = config.DefaultEndpoint
				}
			}
		}
		if err := config.ValidateURL(*endpoint); err != nil {
			return err
		}
		return cli.PrintJSON(a.out, map[string]any{"mcpServers": map[string]any{args[0]: map[string]any{"type": "http", "url": *endpoint}}})
	default:
		return fmt.Errorf("connection must be mcp, acp, raindrop, or executor")
	}
}
