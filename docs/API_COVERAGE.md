# Raindrop API Coverage

This project is REST-first. The MCP endpoint can inform future agent workflows, but the CLI uses the documented REST API for predictable local scripting.

## First-Class Commands

| Area | Commands |
| --- | --- |
| Auth/user | `doctor`, `me`, `user update`, `stats` |
| Raindrops | `list`, `search`, `get`, `add`, `update`, `delete`, `suggest` |
| Bulk raindrops | `batch tag`, `batch move`, `batch delete` |
| Search helpers | `untagged`, `duplicates`, `broken`, `filters` |
| Collections | `collections`, `collection create`, `collection update`, `collection delete`, `collection sort`, `collection expand`, `collection merge`, `collection clean`, `collection empty-trash` |
| Collection covers/icons | `covers`, `covers search-text` |
| Tags | `tags`, `tag rename`, `tag merge`, `tag delete` |
| Highlights | `highlights`, `highlight add`, `highlight update`, `highlight delete` |
| Import helpers | `import parse-url`, `import exists`, `import file` |
| Export | `export --format csv|html|zip` |
| Sharing | `sharing list`, `sharing invite`, `sharing role`, `sharing remove`, `sharing unshare`, `sharing join` |
| Permanent copy | `cache` |
| Backups | `backups`, `backup create`, `backup download` |
| Multipart uploads | `upload file`, `cover raindrop`, `cover collection` |
| Escape hatch | `raw METHOD PATH [json-body]` |

## Not Yet First-Class

These are intentionally deferred or available through `raw`:

- OAuth application flows.
- Any undocumented response fields.

## Parity Status

The CLI has first-class coverage for the documented REST endpoint families surfaced in the docs: user, stats, raindrops, batch raindrops, filters, collections, collection covers/icons, tags, highlights, import helpers, export, sharing, cache, backups, and multipart uploads.

`raw` remains useful for endpoint experiments and newly released API capabilities. It is JSON-only by design; use the first-class upload commands for multipart endpoints.

## Parity Strategy

1. Add stable documented endpoints as readable first-class commands.
2. Keep `raw` for full JSON endpoint reach while the command surface matures.
3. Prefer `--json` for agent workflows and tab-separated output for shell workflows.
4. Add dry-run planning before advanced organizer commands mutate many bookmarks.
