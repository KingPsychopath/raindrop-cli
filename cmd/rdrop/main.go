package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"

	"github.com/KingPsychopath/raindrop-cli/internal/cli"
	"github.com/KingPsychopath/raindrop-cli/internal/config"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type app struct {
	out  io.Writer
	err  io.Writer
	json bool
	api  *raindrop.Client
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "rdrop:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	a := &app{out: os.Stdout, err: os.Stderr}
	return a.run(ctx, args)
}

func (a *app) run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.usage()
		return nil
	}
	if args[0] == "help" {
		return a.help(args[1:])
	}
	if args[0] == "--help" || args[0] == "-h" {
		a.usage()
		return nil
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Fprintln(a.out, versionString())
		return nil
	}
	if args[0] == "completion" {
		return a.completion(args[1:])
	}
	if args[0] == "config" {
		return a.config(args[1:])
	}
	if args[0] == "connect" {
		return a.connect(args[1:])
	}
	if args[0] == "organize" {
		for _, arg := range args[1:] {
			if isHelpArg(arg) {
				return a.help([]string{"organize"})
			}
		}
		return a.organizeCommand(args[1:])
	}
	if args[0] == "mcp" {
		return a.mcp(ctx, args[1:])
	}
	for i, arg := range args[1:] {
		if isHelpArg(arg) {
			return a.help(args[:i+1])
		}
		if strings.HasPrefix(arg, "-") {
			break
		}
	}
	if len(args) > 1 && isHelpArg(args[1]) {
		return a.help([]string{args[0]})
	}

	if a.api == nil {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		a.api = raindrop.NewClient(cfg.APIBaseURL, cfg.Token)
	}
	return a.dispatch(ctx, args)
}

func (a *app) dispatch(ctx context.Context, args []string) error {

	switch args[0] {
	case "doctor":
		if err := noArguments("doctor", args[1:]); err != nil {
			return err
		}
		return a.doctor(ctx)
	case "raw":
		return a.raw(ctx, args[1:])
	case "me":
		if err := jsonOnly("me", args[1:]); err != nil {
			return err
		}
		raw, err := a.api.User(ctx)
		return a.printRaw(raw, err)
	case "user":
		return a.user(ctx, args[1:])
	case "stats":
		if err := jsonOnly("stats", args[1:]); err != nil {
			return err
		}
		stats, err := a.api.Stats(ctx)
		return a.print(stats, err)
	case "collections":
		return a.collections(ctx, args[1:])
	case "collection":
		return a.collection(ctx, args[1:])
	case "export":
		return a.export(ctx, args[1:])
	case "import":
		return a.importCmd(ctx, args[1:])
	case "upload":
		return a.upload(ctx, args[1:])
	case "cover":
		return a.cover(ctx, args[1:])
	case "tags":
		return a.tags(ctx, args[1:])
	case "tag":
		return a.tag(ctx, args[1:])
	case "highlights":
		return a.highlights(ctx, args[1:])
	case "highlight":
		return a.highlight(ctx, args[1:])
	case "cache":
		return a.cache(ctx, args[1:])
	case "sharing":
		return a.sharing(ctx, args[1:])
	case "covers":
		return a.covers(ctx, args[1:])
	case "backups":
		return a.backups(ctx, args[1:])
	case "backup":
		return a.backup(ctx, args[1:])
	case "filters":
		return a.filters(ctx, args[1:])
	case "cleanup":
		return a.cleanup(ctx, args[1:])
	case "context":
		return a.contextCommand(ctx, args[1:])
	case "classify":
		return a.classifyCommand(ctx, args[1:])
	case "agent":
		return a.agent(ctx, args[1:])
	case "list", "search", "untagged", "duplicates", "broken":
		return a.list(ctx, args[0], args[1:])
	case "get":
		return a.get(ctx, args[1:])
	case "add":
		return a.add(ctx, args[1:])
	case "update":
		return a.update(ctx, args[1:])
	case "delete":
		return a.delete(ctx, args[1:])
	case "batch":
		return a.batch(ctx, args[1:])
	case "suggest":
		return a.suggest(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func versionString() string {
	return fmt.Sprintf("%s commit:%s date:%s", version, commit, date)
}

func (a *app) usage() {
	fmt.Fprintln(a.out, `rdrop: Raindrop.io CLI, MCP tools and model-driven organizing.

Usage: rdrop <command> [flags]

Read:     doctor me stats collections tags filters list search get
Organize: cleanup context organize classify agent
Write:    add update delete batch collection tag highlight sharing user
Transfer: export import upload cover cache backups backup
Connect:  mcp connect config
Other:    highlights covers suggest completion version help raw

Start with: rdrop cleanup report --json
MCP:        rdrop mcp serve                 (read-only by default)
Model loop: rdrop agent --model <model> --json <task>
Reference:  rdrop help <command>            (all flags and behavior)
            rdrop help --markdown

TSV list output by default; --json for structured data. Exports write bytes.
Bulk writes need an explicit source collection. Review saved plans before apply.`)
}

func isHelpArg(arg string) bool {
	return arg == "help" || arg == "--help" || arg == "-h"
}

func (a *app) help(args []string) error {
	if len(args) == 0 {
		a.usage()
		return nil
	}
	if len(args) == 1 && args[0] == "--markdown" {
		fmt.Fprintln(a.out, "# Command reference\n\nGenerated by `rdrop help --markdown`. Flags may precede or follow positional arguments. Use `--` before values that begin with a dash.")
		keys := make([]string, 0, len(helpTopics))
		for key := range helpTopics {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintln(a.out)
			fmt.Fprintf(a.out, "## %s\n\n```text\n%s\n```\n", key, strings.TrimSpace(helpTopics[key]))
		}
		return nil
	}
	topic := strings.Join(args, " ")
	if text, ok := helpTopics[topic]; ok {
		fmt.Fprintln(a.out, strings.TrimSpace(text))
		return nil
	}
	if text, ok := helpTopics[args[0]]; ok {
		fmt.Fprintln(a.out, strings.TrimSpace(text))
		return nil
	}
	return fmt.Errorf("unknown help topic %q", topic)
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		f := fs.Lookup(name)
		if f == nil || isBoolFlag(f) || strings.Contains(arg, "=") {
			continue
		}
		if i+1 >= len(args) {
			continue
		}
		i++
		flags = append(flags, args[i])
	}
	return fs.Parse(append(flags, positionals...))
}

func isBoolFlag(f *flag.Flag) bool {
	type boolFlag interface {
		IsBoolFlag() bool
	}
	value, ok := f.Value.(boolFlag)
	return ok && value.IsBoolFlag()
}

func (a *app) completion(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: rdrop completion bash|zsh|fish")
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(a.out, bashCompletion)
	case "zsh":
		fmt.Fprint(a.out, zshCompletion)
	case "fish":
		fmt.Fprint(a.out, fishCompletion)
	default:
		return fmt.Errorf("usage: rdrop completion bash|zsh|fish")
	}
	return nil
}

func (a *app) print(value any, err error) error {
	if err != nil {
		return err
	}
	return cli.PrintJSON(a.out, value)
}

func (a *app) printRaw(raw json.RawMessage, err error) error {
	if err != nil {
		return err
	}
	var value any
	if json.Unmarshal(raw, &value) == nil {
		return cli.PrintJSON(a.out, value)
	}
	fmt.Fprintln(a.out, string(raw))
	return nil
}
