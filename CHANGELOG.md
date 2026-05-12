# Changelog

All notable project changes are documented here.

This project uses semantic version tags for releases. Dates are in UTC.

## Unreleased

### Added

- `scripts/render-terminal-demo.go` and `make assets` for regenerating the README terminal SVG.
- Build metadata injection for `rdrop version`.
- Install and checksum verification examples in the README.
- Code of conduct.

### Changed

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
