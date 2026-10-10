# Official MCP coverage

`rdrop` combines local REST tools and reviewed plans with an optional connection to [Raindrop's official MCP service](https://help.raindrop.io/integrations/mcp). This covers the official AI read workflows without claiming that local REST search implements Raindrop's semantic index.

## Connect the official reads

The normal private Raindrop token configuration is reused. No token is printed or included in generated client files.

```sh
rdrop mcp tools --official --search 'semantic search'
rdrop mcp tools --official --name raindrop_official_find_bookmarks
rdrop mcp serve --official
rdrop agent --official --model your-tool-capable-model --json \
  'Find saved links about software architecture, review their current tags, and propose a plan.'
```

`--official` connects to the fixed `https://api.raindrop.io/rest/v2/ai/mcp` endpoint through Streamable HTTP and imports the nine documented read tools under `raindrop_official_`. The live schemas come from the service; discovery with this flag requires authentication and network access. Default discovery stays offline, and default commands use REST without that service. Initialization errors fail explicitly rather than silently substituting a different search.

Read the schema before making a direct call. For example, save `{"search":"software architecture","limit":5}` in a private `arguments.json`, then:

```sh
rdrop mcp call --official --name raindrop_official_find_bookmarks --input arguments.json
```

Native REST list pages are zero-based and capped at 50. The official `find_bookmarks` schema uses one-based pages and a different limit; inspect its current schema. Do not translate pagination or result shapes by guessing.

The official service restricts text and semantic search on the free plan; full access requires Pro. Other upstream availability, limits and errors are passed through. `rdrop` cannot unlock subscription features. All nine read calls and schema discovery were verified against the live service; empty audit suggestions do not prove classification quality. [Verification evidence](VERIFICATION.md).

## Workflow comparison

| Official workflow | In `rdrop` |
| --- | --- |
| Find bookmarks, including semantic/text search | `list`/`search` for REST queries; `raindrop_official_find_bookmarks` for hosted semantic/text search |
| Fetch bookmark content | `get` for metadata, `cache` for exported cached bytes; `raindrop_official_fetch_bookmark_content` for hosted extracted content |
| Find misplaced bookmarks | `raindrop_official_find_misplaced_bookmarks`; review its suggestions through local context/plans |
| Find mistagged bookmarks | `raindrop_official_find_mistagged_bookmarks`; local `classify` also supports a caller-defined taxonomy |
| Create/update/delete bookmarks | Local `add`, `update`, `delete`; reviewed `batch add` plus explicit bulk tag/move/delete plans |
| Find/create/update/merge/delete collections | Local collection reads and writes; `raindrop_official_find_collections` for hosted name search |
| Find/update/delete tags | Local tag list/rename/merge/delete; `raindrop_official_find_tags` for hosted tag search |
| Find/create/update/delete highlights | Local highlight reads and writes; `raindrop_official_find_highlights` for hosted search |
| Current account | Local `me`/`stats`; `raindrop_official_fetch_current_user` |
| Popular keywords | `raindrop_official_fetch_popular_keywords` |

Local mutations deliberately keep their documented CLI shapes and safety rules. `batch add` supports 1..100 HTTP(S) bookmarks with link, title, tags, note, excerpt, collection, favorite and metadata-parsing fields. Preview and review an exact creation plan before applying it. The optional bridge never imports upstream mutations, even when local writes are enabled. Equivalent workflows do not imply identical tool schemas or transport features.

```sh
rdrop batch add --file new-bookmarks.json --dry-run > reviewed-plan.json
rdrop batch add --plan reviewed-plan.json --dry-run
rdrop batch add --plan reviewed-plan.json --yes
```

The input file is an array, for example `[{"link":"https://example.com","tags":["reference"],"collection":{"$id":-1}}]`. MCP uses `raindrop_add_batch_preview` and the explicitly enabled `raindrop_add_batch_apply`. The plan binds the exact payload and account and checks destination collections. Creation does not have existing-bookmark snapshots. If a request fails or returns incomplete results, inspect the saved URLs before retrying; the API provides no transaction or guaranteed idempotency.

This project additionally provides shell exports, tree paths, bounded model execution, provider-neutral proposals, exact snapshot review/recovery, local decision providers and native packaging. These are concrete workflow differences; tool count alone is not a quality or performance comparison.

## Security and hosting

The bridge binds the credential to the configured official origin, refuses redirects, limits response/event sizes, disables automatic reconnect retries, and imports a fixed allowlist of documented reads. A remote tool annotation cannot grant mutation access. Local input validation runs before forwarding a call. AI suggestions remain untrusted proposals.

This is an outbound MCP client inside the local server/agent. It does not turn `rdrop` into a public OAuth HTTP server for browser-only clients. Use the [client guide](CLIENTS.md) to choose the correct connection.
