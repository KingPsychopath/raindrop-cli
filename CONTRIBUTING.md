# Contributing

Thanks for helping make `raindrop-cli` useful and boring in the best way.

## Setup

```bash
git clone https://github.com/KingPsychopath/raindrop-cli.git
cd raindrop-cli
go test ./...
make build
```

For live API testing:

```bash
export RAINDROP_TOKEN=your_token_here
./bin/rdrop doctor
```

## Pull Request Checklist

- Keep command names clear and consistent with existing verbs.
- Add or update tests for parsing, client behavior, or command behavior when possible.
- Update `README.md` and `docs/API_COVERAGE.md` when adding a first-class command.
- Run `make check`.
- Do not include personal exports, tokens, or generated binaries.
- Keep bulk destructive workflows explicit with `--dry-run`, `--yes`, or ID-based inputs.

## CLI Design

- Human output should be compact and pipeline-friendly.
- Machine output should use `--json`.
- Commands that emit file formats, such as `export`, write the payload to stdout so users can redirect or pipe it.
- Dangerous bulk operations should require targeted inputs such as `--ids`, `--search`, or a collection ID.

## Releases

Maintainers publish releases by pushing a semantic version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds Linux, macOS, and Windows archives for amd64 and arm64.
