# Roadmap

This roadmap describes possible directions for `raindrop-cli`. It is not a schedule, promise, or commitment. There are no expected delivery dates for any item here.

The project should stay small, scriptable, dependency-light, and safe for large Raindrop.io libraries.

## Completed

- First-class REST commands for core Raindrop.io workflows.
- Tab-separated default output for shell pipelines.
- JSON output for structured workflows.
- Export and backup downloads written directly to stdout.
- Cleanup report for safe organizing workflows.
- Explicit guardrails for bulk delete and empty-trash operations.
- Shell completions for Bash, Zsh, and Fish.
- GitHub Actions CI.
- Tagged releases with Linux, macOS, and Windows binaries for amd64 and arm64.
- SHA-256 checksums for release artifacts.
- Changelog for human-readable release history.
- Release build metadata in `rdrop version`.
- Code of conduct.
- Regeneratable terminal demo SVG.
- Dependabot for GitHub Actions and Go module updates.
- Homebrew tap for source-based installs.

## Near-Term Ideas

These are useful next steps, but they have no expected timing.

- Improve `rdrop help` with grouped command sections.
- Add `--format tsv|json` consistently where it makes sense.
- Add more install snippets for package managers if packaging expands.
- Add a `rdrop config paths` or `rdrop config doctor` command for auth/debugging.
- Add examples for common Raindrop search syntax.
- Add command-specific docs under `docs/commands/` if the README gets too large.

## Safety And Cleanup

These items should remain conservative and reviewable.

- Add apply/review files for cleanup plans.
- Add duplicate review helpers that group likely duplicates without deleting automatically.
- Add tag review helpers for near-duplicate tags.
- Add `--dry-run` planning for more bulk workflows where useful.
- Add better before/after summaries for cleanup runs.
- Keep destructive commands explicit and easy to audit.

## Release And Packaging

These are packaging improvements to consider over time.

- Use friendlier artifact names such as `macos` while keeping Go's internal `darwin` build target.
- Consider Homebrew bottles if there is demand.
- Add Scoop or Winget support for Windows if there is demand.
- Add Debian/RPM packages only if the extra maintenance is justified.
- Consider signed macOS binaries if distribution friction becomes a real problem.
- Keep checksums published for every release.

## API Coverage

`rdrop raw` should remain available for parity gaps. First-class commands should be added when an endpoint is stable and the command UX is clear.

Possible future API work:

- Track newly documented Raindrop.io REST endpoints.
- Promote frequently used `raw` workflows into first-class commands.
- Improve multipart upload ergonomics.
- Add richer sharing and collaboration output.
- Improve permanent-copy/cache workflows for Pro accounts.

## Not Planned By Default

These are intentionally not default goals unless the project direction changes.

- A long-running daemon.
- A web UI.
- A database cache.
- Automatic destructive cleanup without a review step.
- Heavy dependencies for basic command parsing or output.

## Contributing To The Roadmap

Open an issue with the workflow you want, the command shape you imagine, and the safety implications. For destructive or bulk operations, include the dry-run or review path you would expect.
