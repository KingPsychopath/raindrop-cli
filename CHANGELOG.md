# Changelog

All notable project changes are documented here.

This project uses semantic version tags for releases. Dates are in UTC.

## v0.2.0 - 2026-10-10

### Added

- Official Go SDK MCP stdio server, structured tool catalog, read-only defaults, prompt/resource and mutation intent checks.
- Bounded direct-model Responses/Chat tool loop with exact plan capture and optional HTTP/local-command decisions, including Jev and Laya adapters.
- Provider-neutral library context, existing naming policy, multi-tag/move proposal validation and reviewed account/snapshot-bound apply with partial recovery.
- ACP session configuration, official Raindrop/Executor MCP connection helpers and a portable Executor app with exact-call host approval.
- Native npm launcher/platform build, pack and integrity-checked publish scripts, non-root Dockerfile and generated command reference.
- Repeatable real Qwen/Laya verification against a disposable library and local Executor setup helpers.

- `scripts/render-terminal-demo.go` and `make assets` for regenerating the README terminal SVG.
- Build metadata injection for `rdrop version`.
- Install and checksum verification examples in the README.
- Code of conduct.
- Dependabot configuration.
- Live Homebrew tap instructions and packaging notes.

### Changed

- Split command implementation into focused modules shared with model tools.
- Use the source-declared Go version in CI/release builds; validate pagination, explicit bulk source collections, structured inputs and extra arguments.
- Add dry-run tag/move snapshot plans, preserve existing tags on proposal apply, redact diagnostics and block cross-origin credential redirects.
- Isolate each classified bookmark in its own provider state after real local inference exposed interference between unrelated items.
- Bind Executor credentials to the Raindrop API host and add a safe account health check.
- Restore the README status/version badges and update the terminal graphic.
- Rewrite command, integration, model, organizing and packaging guides.

- Future macOS release artifacts use `macos` in filenames instead of Go's internal `darwin` target name.

## v0.1.0 - 2026-05-12

Initial open-source release.

### Added

- Dependency-light Go CLI for Raindrop.io REST workflows.
- Commands for bookmarks, collections, tags, highlights, filters, imports, exports, backups, sharing, and raw JSON API calls.
- `cleanup report` for read-only organizing workflows.
- Bulk safety guardrails for delete and empty-trash flows.
- Bash, Zsh, and Fish completion output.
- GitHub Actions CI.
- GitHub release workflow with Linux, macOS, and Windows binaries for amd64 and arm64.
- SHA-256 checksums for release assets.
- README terminal demo image.
