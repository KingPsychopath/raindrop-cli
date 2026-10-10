package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/KingPsychopath/raindrop-cli/internal/cli"
	"net/url"
	"strings"
)

func (a *app) export(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	collection := fs.Int64("collection", 0, "collection id")
	format := fs.String("format", "csv", "csv, html, or zip")
	search := fs.String("search", "", "search query")
	sort := fs.String("sort", "-created", "sort")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if err := noArguments("export", fs.Args()); err != nil {
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
