# raindrop-cli

`rdrop` is a Go CLI for managing large Raindrop.io libraries.

The goal is an agent-friendly, open-source tool for bookmarks, notes, tags, collections, highlights, duplicates, broken links, and cleanup workflows at thousands or tens of thousands of saved items.

The CLI is REST-first because Raindrop's REST API exposes broad command parity today. Raindrop's MCP endpoint is useful as a learning and future intelligence layer, but the command line should remain easy to install, script, audit, and package on every platform.

## Requirements

- Go 1.22 or newer
- A Raindrop.io API token

Raindrop's REST API base URL is:

```text
https://api.raindrop.io/rest/v1
```

## Auth

Set `RAINDROP_TOKEN` in your shell, or put it in:

```text
~/.config/openclaw/gateway.env
```

Example:

```text
RAINDROP_TOKEN=your_token_here
```

Optional overrides:

```text
RAINDROP_API_BASE_URL=https://api.raindrop.io/rest/v1
RAINDROP_MCP_ENDPOINT=https://api.raindrop.io/rest/v2/ai/mcp
```

## Build

```bash
make build
./bin/rdrop doctor
```

To install into your Go bin path:

```bash
make install
rdrop doctor
```

## Commands

```bash
rdrop help
rdrop help cleanup
rdrop doctor
rdrop me
rdrop user update '{"fullName":"New Name"}'
rdrop stats
rdrop collections
rdrop collections --children
rdrop collection create "Reading"
rdrop collection merge --to 123 --ids 456,789
rdrop collection clean
rdrop covers pokemon
rdrop tags
rdrop tag rename old-name new-name
rdrop filters
rdrop list --per-page 50
rdrop search "javascript"
rdrop untagged
rdrop duplicates
rdrop broken
rdrop get 123456
rdrop add --title Example --tags docs,example --parse https://example.com
rdrop update 123456 '{"important":true,"tags":["go","docs"]}'
rdrop delete 123456
rdrop highlights --collection 123
rdrop highlight add 123456 --text "Useful quote" --color yellow
rdrop export --format csv > raindrop.csv
rdrop import exists https://example.com,https://example.org
rdrop import file bookmarks.html
rdrop upload file --collection 123 document.pdf
rdrop cover raindrop 123456 cover.png
rdrop cover collection 123 cover.png
rdrop sharing list 123
rdrop backups
rdrop backup download 659d42a35ffbb2eb5ae1cb86 --format csv > backup.csv
rdrop cleanup report
rdrop cleanup report --json | jq '.tagPlans[]'
rdrop batch tag --ids 1,2,3 --tags needs-review
rdrop batch move --ids 1,2,3 --to 98765
rdrop raw GET user/stats
```

The default list outputs are tab-separated so they work with common tools:

```bash
rdrop search "postgres" | cut -f1,3
rdrop tags | sort -k2 -nr | head
rdrop export --format html > raindrop.html
```

Add `--json` where structured output is useful for agents or `jq`:

```bash
rdrop list --search "postgres" --json | jq '.items[].link'
```

## Development Principles

- Prefer MCP tools for library intelligence and semantic cleanup.
- Keep destructive operations explicit.
- Add `--dry-run` before shipping bulk organizing commands.
- Keep command output useful for both humans and agents.
- Avoid storing tokens in the repository.
- Keep REST parity through first-class commands where the API is stable and through `rdrop raw` everywhere else.

## Project Layout

```text
cmd/rdrop              CLI entrypoint and command parsing
internal/config        Environment and token loading
internal/raindrop      Raindrop REST client and API methods
internal/cli           Formatting and parsing helpers
docs/API_COVERAGE.md   Raindrop endpoint coverage and known gaps
AGENTS.md              Guidance for AI coding agents and contributors
```

## Current Scope

See [`docs/API_COVERAGE.md`](docs/API_COVERAGE.md) for the command parity matrix.

For cleanup and agent workflows, see [`docs/ORGANIZING.md`](docs/ORGANIZING.md).

Next likely additions:

- apply/review files for cleanup plans
- duplicate resolution helpers
- optional MCP-powered organizer commands
