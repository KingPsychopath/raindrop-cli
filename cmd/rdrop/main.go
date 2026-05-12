package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/KingPsychopath/raindrop-cli/internal/cli"
	"github.com/KingPsychopath/raindrop-cli/internal/config"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

const version = "0.1.0"

type app struct {
	out  *os.File
	err  *os.File
	json bool
	api  *raindrop.Client
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "rdrop:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	a := &app{out: os.Stdout, err: os.Stderr}
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
		fmt.Fprintln(a.out, version)
		return nil
	}
	if args[0] == "completion" {
		return a.completion(args[1:])
	}
	if len(args) > 1 && isHelpArg(args[1]) {
		return a.help([]string{args[0]})
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.api = raindrop.NewClient(cfg.APIBaseURL, cfg.Token)

	switch args[0] {
	case "doctor":
		return a.doctor(ctx)
	case "raw":
		return a.raw(ctx, args[1:])
	case "me":
		raw, err := a.api.User(ctx)
		return a.printRaw(raw, err)
	case "user":
		return a.user(ctx, args[1:])
	case "stats":
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

func (a *app) usage() {
	fmt.Fprintln(a.out, `rdrop is a Raindrop.io CLI for large bookmark libraries.

Usage:
  rdrop doctor
  rdrop me [--json]
  rdrop user update '{"fullName":"Name"}'
  rdrop stats [--json]
  rdrop collections [--children] [--json]
  rdrop collection create [--parent id] <title>
  rdrop collection update <id> '{"title":"New name"}'
  rdrop collection delete <id>
  rdrop collection sort title|-title|-count
  rdrop collection expand true|false
  rdrop collection merge --to id --ids id,id
  rdrop collection clean
  rdrop collection empty-trash
  rdrop covers [search-text]
  rdrop export --format csv|html|zip [--collection id] [--search query] [--sort -created] > bookmarks.csv
  rdrop import parse-url <url>
  rdrop import exists <url,url>
  rdrop import file <bookmarks.html>
  rdrop upload file [--collection id] <path>
  rdrop cover raindrop <raindrop-id> <image-path>
  rdrop cover collection <collection-id> <image-path>
  rdrop tags [--collection id] [--json]
  rdrop tag rename [--collection id] <old> <new>
  rdrop tag merge [--collection id] <from,from2> <to>
  rdrop tag delete [--collection id] <tag,tag2>
  rdrop highlights [--collection id] [--page n] [--per-page n]
  rdrop highlight add <raindrop-id> --text text [--color yellow] [--note text]
  rdrop highlight update <raindrop-id> <highlight-id> [--text text] [--color yellow] [--note text]
  rdrop highlight delete <raindrop-id> <highlight-id>
  rdrop cache <raindrop-id>
  rdrop sharing list <collection-id>
  rdrop sharing invite <collection-id> --emails a@b.com,c@d.com [--role viewer|member]
  rdrop sharing role <collection-id> <user-id> --role viewer|member
  rdrop sharing remove <collection-id> <user-id>
  rdrop sharing unshare <collection-id>
  rdrop sharing join <collection-id> --token token
  rdrop backups
  rdrop backup create
  rdrop backup download <backup-id> --format csv|html > backup.csv
  rdrop cleanup report [--collection id] [--search query] [--json]
  rdrop filters [--collection id] [--search query] [--json]
  rdrop list [--collection id] [--search query] [--page n] [--per-page n] [--sort -created] [--nested] [--json]
  rdrop search [list flags] <query>
  rdrop untagged [list flags]
  rdrop duplicates [list flags]
  rdrop broken [list flags]
  rdrop get [--json] <id>
  rdrop add [--title text] [--tags a,b] [--collection id] [--note text] [--excerpt text] [--parse] <url>
  rdrop update <id> '{"tags":["go"],"important":true}'
  rdrop delete <id>
  rdrop batch tag --ids 1,2 --tags go,docs [--collection id]
  rdrop batch move --ids 1,2 --to collectionID [--collection id]
  rdrop batch delete --ids 1,2 [--collection id] --dry-run|--yes
  rdrop suggest <url> [--json]
  rdrop completion bash|zsh|fish
  rdrop raw METHOD PATH [json-body]`)
}

func isHelpArg(arg string) bool {
	return arg == "help" || arg == "--help" || arg == "-h"
}

func (a *app) help(args []string) error {
	if len(args) == 0 {
		a.usage()
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

func isBoolFlag(f *flag.Flag) bool {
	type boolFlag interface {
		IsBoolFlag() bool
	}
	value, ok := f.Value.(boolFlag)
	return ok && value.IsBoolFlag()
}

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

func (a *app) collections(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("collections", flag.ContinueOnError)
	children := fs.Bool("children", false, "include child collections")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	a.json = *jsonOut
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

func (a *app) export(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	format := fs.String("format", "csv", "csv, html, or zip")
	search := fs.String("search", "", "search query")
	sort := fs.String("sort", "-created", "sort")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	switch *format {
	case "csv", "html", "zip":
	default:
		return fmt.Errorf("--format must be csv, html, or zip")
	}
	data, _, err := a.api.Export(ctx, *collection, *format, *search, *sort)
	if err != nil {
		return err
	}
	_, err = a.out.Write(data)
	return err
}

func (a *app) importCmd(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("import subcommand required")
	}
	switch args[0] {
	case "parse-url":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop import parse-url <url>")
		}
		return a.printRaw(a.api.ParseURL(ctx, args[1]))
	case "exists":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop import exists <url,url>")
		}
		return a.printRaw(a.api.URLsExist(ctx, cli.CSV(args[1])))
	case "file":
		if len(args) != 2 {
			return fmt.Errorf("usage: rdrop import file <bookmarks.html>")
		}
		return a.printRaw(a.api.ImportFile(ctx, args[1]))
	default:
		return fmt.Errorf("unknown import subcommand %q", args[0])
	}
}

func (a *app) upload(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("upload subcommand required")
	}
	switch args[0] {
	case "file":
		fs := flag.NewFlagSet("upload file", flag.ContinueOnError)
		collection := fs.Int64("collection", 0, "collection id")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: rdrop upload file [--collection id] <path>")
		}
		return a.printRaw(a.api.UploadRaindropFile(ctx, fs.Arg(0), *collection))
	default:
		return fmt.Errorf("unknown upload subcommand %q", args[0])
	}
}

func (a *app) cover(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("cover subcommand required")
	}
	switch args[0] {
	case "raindrop":
		if len(args) != 3 {
			return fmt.Errorf("usage: rdrop cover raindrop <raindrop-id> <image-path>")
		}
		id, err := cli.Int64(args[1], "raindrop-id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.UploadRaindropCover(ctx, id, args[2]))
	case "collection":
		if len(args) != 3 {
			return fmt.Errorf("usage: rdrop cover collection <collection-id> <image-path>")
		}
		id, err := cli.Int64(args[1], "collection-id")
		if err != nil {
			return err
		}
		return a.printRaw(a.api.UploadCollectionCover(ctx, id, args[2]))
	default:
		return fmt.Errorf("unknown cover subcommand %q", args[0])
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

func (a *app) tags(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("tags", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	a.json = *jsonOut
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
	default:
		return fmt.Errorf("unknown cleanup subcommand %q", args[0])
	}
}

func (a *app) highlights(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("highlights", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	page := fs.Int("page", 0, "page number")
	perPage := fs.Int("per-page", 25, "results per page")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	return a.printRaw(a.api.Highlights(ctx, *collection, *page, *perPage))
}

func (a *app) highlight(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("highlight subcommand required")
	}
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("highlight add", flag.ContinueOnError)
		text := fs.String("text", "", "highlight text")
		color := fs.String("color", "", "highlight color")
		note := fs.String("note", "", "highlight note")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 || *text == "" {
			return fmt.Errorf("usage: rdrop highlight add <raindrop-id> --text text [--color yellow] [--note text]")
		}
		id, err := cli.Int64(fs.Arg(0), "raindrop-id")
		if err != nil {
			return err
		}
		highlight := map[string]any{"text": *text}
		if *color != "" {
			highlight["color"] = *color
		}
		if *note != "" {
			highlight["note"] = *note
		}
		return a.print(a.api.Update(ctx, id, map[string]any{"highlights": []any{highlight}}))
	case "update":
		fs := flag.NewFlagSet("highlight update", flag.ContinueOnError)
		text := fs.String("text", "", "highlight text")
		color := fs.String("color", "", "highlight color")
		note := fs.String("note", "", "highlight note")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 2 {
			return fmt.Errorf("usage: rdrop highlight update <raindrop-id> <highlight-id> [--text text] [--color yellow] [--note text]")
		}
		id, err := cli.Int64(fs.Arg(0), "raindrop-id")
		if err != nil {
			return err
		}
		highlight := map[string]any{"_id": fs.Arg(1)}
		if *text != "" {
			highlight["text"] = *text
		}
		if *color != "" {
			highlight["color"] = *color
		}
		if *note != "" {
			highlight["note"] = *note
		}
		return a.print(a.api.Update(ctx, id, map[string]any{"highlights": []any{highlight}}))
	case "delete":
		if len(args) != 3 {
			return fmt.Errorf("usage: rdrop highlight delete <raindrop-id> <highlight-id>")
		}
		id, err := cli.Int64(args[1], "raindrop-id")
		if err != nil {
			return err
		}
		return a.print(a.api.Update(ctx, id, map[string]any{
			"highlights": []any{map[string]any{"_id": args[2], "text": ""}},
		}))
	default:
		return fmt.Errorf("unknown highlight subcommand %q", args[0])
	}
}

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
	dryRun := fs.Bool("dry-run", false, "print the planned bulk delete without mutating")
	yes := fs.Bool("yes", false, "confirm a destructive bulk delete")
	jsonOut := fs.Bool("json", false, "print JSON for dry-run output")
	if err := parseFlags(fs, args[1:]); err != nil {
		return err
	}
	ids, err := cli.Int64CSV(*idsFlag)
	if err != nil {
		return err
	}
	switch args[0] {
	case "tag":
		if len(ids) == 0 || len(cli.CSV(*tagsFlag)) == 0 {
			return fmt.Errorf("usage: rdrop batch tag --ids 1,2 --tags go,docs [--collection id]")
		}
		body := map[string]any{"tags": cli.CSV(*tagsFlag)}
		body["ids"] = ids
		return a.printRaw(a.api.BatchUpdate(ctx, *collection, body))
	case "move":
		if len(ids) == 0 || *to == 0 {
			return fmt.Errorf("usage: rdrop batch move --ids 1,2 --to collectionID [--collection id]")
		}
		body := map[string]any{"collection": raindrop.Ref{ID: *to}, "ids": ids}
		return a.printRaw(a.api.BatchUpdate(ctx, *collection, body))
	case "delete":
		if len(ids) == 0 && *search == "" {
			return fmt.Errorf("usage: rdrop batch delete --ids 1,2|--search query [--collection id] --dry-run|--yes")
		}
		if *dryRun {
			return a.printBatchDeletePlan(ctx, *collection, *search, ids, *jsonOut)
		}
		if !*yes {
			return fmt.Errorf("refusing bulk delete without --dry-run or --yes")
		}
		return a.printRaw(a.api.BatchDelete(ctx, *collection, *search, ids))
	default:
		return fmt.Errorf("unknown batch subcommand %q", args[0])
	}
}

func (a *app) printBatchDeletePlan(ctx context.Context, collectionID int64, search string, ids []int64, jsonOut bool) error {
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

func (a *app) suggest(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: rdrop suggest <url>")
	}
	return a.printRaw(a.api.Suggest(ctx, args[0]))
}

func (a *app) cache(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: rdrop cache <raindrop-id>")
	}
	id, err := cli.Int64(args[0], "raindrop-id")
	if err != nil {
		return err
	}
	data, _, err := a.api.Cache(ctx, id)
	if err != nil {
		return err
	}
	_, err = a.out.Write(data)
	return err
}

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

func (a *app) backups(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: rdrop backups")
	}
	return a.printRaw(a.api.Backups(ctx))
}

func (a *app) backup(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("backup subcommand required")
	}
	switch args[0] {
	case "create":
		if len(args) != 1 {
			return fmt.Errorf("usage: rdrop backup create")
		}
		data, _, err := a.api.CreateBackup(ctx)
		if err != nil {
			return err
		}
		_, err = a.out.Write(data)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			fmt.Fprintln(a.out)
		}
		return err
	case "download":
		fs := flag.NewFlagSet("backup download", flag.ContinueOnError)
		format := fs.String("format", "csv", "csv or html")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: rdrop backup download <backup-id> --format csv|html")
		}
		if *format != "csv" && *format != "html" {
			return fmt.Errorf("--format must be csv or html")
		}
		data, _, err := a.api.DownloadBackup(ctx, fs.Arg(0), *format)
		if err != nil {
			return err
		}
		_, err = a.out.Write(data)
		return err
	default:
		return fmt.Errorf("unknown backup subcommand %q", args[0])
	}
}

func (a *app) raw(ctx context.Context, args []string) error {
	if len(args) < 2 || len(args) > 3 {
		return fmt.Errorf("usage: rdrop raw METHOD PATH [json-body]")
	}
	var body any
	if len(args) == 3 {
		parsed, err := cli.JSONObject(args[2])
		if err != nil {
			return err
		}
		body = parsed
	}
	return a.printRaw(a.api.Raw(ctx, strings.ToUpper(args[0]), args[1], url.Values{}, body))
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
