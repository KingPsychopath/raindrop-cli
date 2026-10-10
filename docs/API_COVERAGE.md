# REST and MCP coverage

Stable endpoints follow the [official Raindrop REST documentation](https://developer.raindrop.io/). `rdrop raw` remains the CLI escape hatch for newly documented JSON endpoints; it is deliberately absent from MCP. This table describes the development source, not old release artifacts.

| Area | CLI | Local MCP / direct loop |
| --- | --- | --- |
| Account | `doctor`, `me`, `user update`, `stats` | User, stats; update with write opt-in |
| Bookmarks | `list`, `search`, `get`, `add`, `update`, `delete`, `suggest` | Paged list with search, get, suggest; writes with opt-in |
| Search helpers | `untagged`, `duplicates`, `broken`, `filters` | Same structured reads |
| Bulk bookmarks | `batch tag`, `batch move`, `batch delete` | Exact-ID tag/move/delete previews; saved cleanup apply with write opt-in |
| Organizing | `cleanup report`, `cleanup plan/apply`, `context`, `organize plan` | Report, context, validated multi-tag/move plan, saved apply |
| Typed decisions | `classify` | `raindrop_classify`, configured optional provider |
| Collections | Roots, children, `--tree`; create/update/delete/sort/expand/merge/clean/empty-trash | Same JSON operations; writes with opt-in |
| Tags | List, rename/merge/delete | Same; writes with opt-in |
| Highlights | Paged list, add/update/delete | Same; writes with opt-in |
| Cover search | `covers` | Same |
| Import helpers | `import parse-url`, `import exists`, `import file` | URL parsing and existence only |
| Sharing | List/invite/role/remove/unshare/join | Same JSON operations; writes with opt-in |
| Backups | List/create/download | List and create; create queues an export/email with opt-in |
| Bytes and local files | Export, cache, backup download, import file, upload file, cover upload | CLI only |
| Raw JSON API | `raw METHOD PATH [json-body]` | CLI only |
| Connections | `connect mcp/acp/raindrop/executor`, `config` | Offline setup helpers; not remote tool calls |
| Direct model loop | `agent` | Uses the same local catalog; never recursively exposed as a tool |

`rdrop mcp tools` is the authoritative read-only tool list; `rdrop mcp tools --allow-write` includes writes. Each tool carries an input schema and annotations; the handler validates the schema before dispatching to the shared command code. Full inputs and CLI flags are in [COMMANDS.md](COMMANDS.md).

## Limits and semantics

Raindrop's documented [multiple-bookmark endpoints](https://developer.raindrop.io/v1/raindrops/multiple) limit pages to 50 and multi-create to 100. This CLI limits explicit bulk IDs and saved manual plans to 100, contexts to 50 and decision calls to 25. It validates duplicate IDs and page bounds locally. Collection `0` is for reading all items except Trash; bulk updates/deletes need a real collection, `-1` or `-99`.

Tag append plans preserve existing names because individual Raindrop updates replace the `tags` array. Direct bulk append uses the documented bulk behavior. Context/proposal planning supports exact tag strings including punctuation, without CSV conversion; CLI `--tags` and tag-merge commands use comma-separated values.

OAuth app provisioning and a public authenticated MCP HTTP server are not implemented. A test token is sufficient for the local REST-backed server; some API operations such as permanent copies are Pro-only. The separate [official Raindrop remote MCP](https://developer.raindrop.io/mcp/mcp) has its own beta/Pro and auth requirements.

## Output and failure contracts

Lists default to TSV with terminal control characters sanitized. JSON preserves metadata. File transfers write bytes. MCP returns structured JSON and tool-error results on failures. API `result:false` is a failure even on HTTP 200; HTTP 429 includes retry information. Redirects cannot take the Raindrop credential to another origin. Mutations are not automatically retried.

The Executor adapter is intentionally smaller: cleanup counts, paged list/get, roots/children, tags/highlights, add and append-tags. Its public framework-host contract is tested independently. It does not implement the Go plan engine or claim full MCP parity.
