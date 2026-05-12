# Raindrop API Coverage

`raindrop-cli` is REST-first. The CLI wraps stable documented Raindrop.io REST endpoints with scriptable commands and keeps `rdrop raw` for parity gaps and newly released JSON endpoints.

## Coverage Summary

| Area | Commands |
| --- | --- |
| Auth and account | `doctor`, `me`, `user update`, `stats` |
| Bookmarks | `list`, `search`, `get`, `add`, `update`, `delete`, `suggest` |
| Bulk bookmarks | `batch tag`, `batch move`, `batch delete` |
| Search helpers | `untagged`, `duplicates`, `broken`, `filters` |
| Collections | `collections`, `collection create`, `collection update`, `collection delete`, `collection sort`, `collection expand`, `collection merge`, `collection clean`, `collection empty-trash` |
| Covers and icons | `covers`, `cover raindrop`, `cover collection` |
| Tags | `tags`, `tag rename`, `tag merge`, `tag delete` |
| Highlights | `highlights`, `highlight add`, `highlight update`, `highlight delete` |
| Import helpers | `import parse-url`, `import exists`, `import file` |
| Export | `export --format csv\|html\|zip` |
| Sharing | `sharing list`, `sharing invite`, `sharing role`, `sharing remove`, `sharing unshare`, `sharing join` |
| Permanent copy | `cache` |
| Backups | `backups`, `backup create`, `backup download` |
| Multipart uploads | `upload file`, `cover raindrop`, `cover collection` |
| Escape hatch | `raw METHOD PATH [json-body]` |

## Output Behavior

Commands use one of three output shapes:

- Tab-separated rows for human and shell list output.
- Pretty JSON when `--json` is available.
- Raw bytes to stdout for exports, backups, and permanent-copy downloads.

This keeps commands friendly to pipelines:

```bash
rdrop tags | sort -k2 -nr | head
rdrop list --json | jq '.items[].link'
rdrop export --format csv > bookmarks.csv
```

## Safety Behavior

The command surface intentionally avoids accidental broad mutation:

- `cleanup report` is read-only.
- `batch tag` and `batch move` require `--ids`.
- `batch delete` requires `--dry-run` or `--yes`.
- `collection empty-trash` requires `--yes`.
- `raw` is JSON-only; multipart endpoints use first-class commands.

## Not First-Class

These are intentionally deferred or available through `raw`:

- OAuth application flows.
- Newly released JSON endpoints before they receive command UX.
- Undocumented response fields.

## Parity Strategy

1. Add stable documented endpoints as readable first-class commands.
2. Keep `raw` available for JSON endpoint reach while the command surface matures.
3. Prefer `--json` for agent workflows and tab-separated output for shell workflows.
4. Add dry-run planning before advanced organizer commands mutate many bookmarks.
5. Keep dependencies minimal unless a dependency materially improves maintainability.
