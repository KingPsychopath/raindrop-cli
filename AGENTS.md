# AGENTS.md

Guidance for AI coding agents and human contributors working in this repository.

## Project Intent

`raindrop-cli` is a dependency-light Go CLI, MCP server and model-driven organizer for managing Raindrop.io libraries at scale. It should stay easy to audit, easy to package, and friendly to shell pipelines.

## Development Rules

- Prefer the Go standard library unless a dependency clearly improves maintainability.
- Keep commands scriptable: plain tab-separated output by default, `--json` where structured data is useful, and exported bytes written directly to stdout.
- Keep destructive commands explicit. Bulk delete and cleanup workflows should grow `--dry-run` planning before becoming more automated.
- Use `rdrop cleanup report --json` as the first step for organizing workflows. Do not mutate a user's library until a plan is visible and intentionally applied.
- Do not commit tokens, exports, local `.env` files, or generated binaries.
- For Go behavior changes, start with focused package tests; run `make check`
  before submitting source changes. Prose-only edits need reference and diff
  review, not application tests.

## Code Map

- `cmd/rdrop`: command parsing and command UX.
- `internal/raindrop`: REST client and Raindrop API methods.
- `internal/cli`: parsing and output helpers.
- `internal/config`: environment and token loading.
- `internal/organize`: versioned context, proposals, snapshot plans and apply recovery.
- `internal/ai`: bounded model HTTP/command adapters and tool loops.
- `cmd/rdrop/protocol_catalog.go`: structured tool schemas sharing command handlers.
- `integrations`: optional Executor and Laya adapters.
- `packaging/npm`, `scripts/build-npm.mjs`: launcher and native platform packaging.
- `docs`: contributor-facing design and API coverage notes.

- Keep model/provider selection outside tool inputs. Context must preserve user tags and collection paths; uncertain suggestions stay out of apply plans.
- MCP and direct model loops are read-only by default. Host authorization is separate from tool confirmation fields.
- Regenerate `docs/COMMANDS.md` with `make docs` after changing help.

## Raindrop API Notes

Use the official Raindrop documentation as the source of truth. First-class commands should wrap stable REST endpoints. Keep `rdrop raw` available for parity gaps, experiments, and newly released endpoints.
