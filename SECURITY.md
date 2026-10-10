# Security boundaries

Report security issues privately to the maintainer before posting account data or a proof of concept publicly.

## Credentials and data

Tokens come from environment variables, explicitly selected private files, the native config path or the legacy OpenClaw fallback. `config show` redacts tokens. Env files are literal values, never executed. Generated connection configs omit credentials. Keep config files and saved contexts/plans private; full snapshots include notes. Do not commit tokens, personal exports, local `.env` files or binaries.

The Raindrop client refuses redirects outside its configured origin. Model clients reject all redirects, require HTTPS for remote endpoints and bound responses. A configured model receives library metadata and may retrieve more through read tools; choose it deliberately. Text from titles, notes, links or model output is data, not shell instructions. TSV strips terminal controls; JSON retains original values.

A local command decision provider is trusted executable code, selected by the operator rather than a model. It is invoked without a shell and standard API-token variables are stripped, but it retains normal local permissions and other inherited environment variables. Do not treat this as a sandbox.

## Permissions and writes

Local MCP and direct-model execution are read-only unless `--allow-write` is supplied. That flag explicitly exposes mutations, including destructive account/library operations. A tool's `confirm:true` proves only caller intent; MCP host review must be enforced independently. Executor's adapter uses its actual exact-call approval mechanism for writes.

Saved cleanup plans bind account and snapshots and detect changed data before the first write. SHA-256 digests detect changed files, not forgery by an authorized local caller. Raindrop lacks a transaction/conditional update for this flow, so concurrent edits can race and partial writes can happen. Inspect uncertain outcomes and create fresh plans; mutations are never automatically retried.

Direct legacy CLI update/delete/tag/collection commands remain immediate writes. `raw` can mutate any JSON endpoint within the configured API origin. Search-based bulk delete is dynamic, and Trash deletion can be permanent. Prefer explicit IDs and reviewed plans for organizing.

## Distribution

Release checksums and npm binary integrity are checked artifact identities, not protection against a compromised publisher. Pin versions and obtain checksums from the corresponding trusted release. The local MCP server has no HTTP listener or built-in OAuth boundary. A remote deployment must provide its own authentication, account isolation and authorization.
