# Agent and tool connections

| Path | Who owns the model loop | Transport / boundary |
| --- | --- | --- |
| `rdrop mcp serve` | Existing MCP client/agent | Local stdio, official Go MCP SDK |
| `rdrop agent` | `rdrop` | Responses or Chat Completions HTTP; shared validated tools |
| `rdrop connect acp` | Existing ACP agent and its client | Generated `session/new` params containing this MCP server |
| Official Raindrop MCP | Remote agent/client | Raindrop's Streamable HTTP service and auth |
| Executor adapter | Executor host/client | Portable TypeScript REST app with account binding and exact-call write approval |
| Context/proposal files | Any model, process or rule engine | Versioned JSON; offline proposal validation, explicit snapshot apply |

## Local MCP

Install the binary and run `rdrop connect mcp --command /absolute/path/rdrop`. Copy its configuration into a client that accepts `mcpServers`:

```json
{"mcpServers":{"raindrop-cli":{"command":"/absolute/path/rdrop","args":["mcp","serve"]}}}
```

The client process must have access to the private config file or inherit `RAINDROP_TOKEN`; GUI clients do not necessarily inherit shell variables. Generated configs omit tokens. `rdrop config show` diagnoses the CLI's resolved paths without printing credentials.

`rdrop mcp tools` prints the read-only catalog without credentials or network calls. `--allow-write` adds mutation tools, including `raindrop_cleanup_apply`; they require `confirm:true`. This field is caller intent, not an approval challenge. Use the host's actual authorization controls. Read-only mode still supports snapshot previews and `raindrop_organize_plan`.

The server has an `organize_library` prompt and `rdrop://guides/organizing` resource. Begin with `raindrop_cleanup_report` and `raindrop_context`. All input schemas are validated locally, including unknown fields, bounds and write confirmation. Command arguments are constructed directly, without a shell. Stdout carries only MCP messages. Errors use MCP tool-error results rather than pretending that API failure was a successful action.

See [coverage](API_COVERAGE.md) for exclusions. Raw HTTP, local file uploads and binary exports remain CLI-only. The MCP server has no public HTTP listener or OAuth server; use a host with its own authentication if you need remote access.

## ACP

[ACP](https://agentclientprotocol.com/protocol/v1/session-setup) is the client-to-agent protocol. This project supplies tools to an existing ACP agent rather than replacing that agent:

```sh
rdrop connect acp --cwd "$PWD" --command /absolute/path/rdrop
```

Output is the params object for `session/new`, after your client initializes a compatible agent. It contains an absolute cwd and a stdio `mcpServers` entry with `name`, `command`, `args`, `env:[]`. The selected agent must advertise/support MCP servers. Its handling of inherited environment and permissions determines token access. `rdrop` itself does not speak ACP session or prompt methods.

## Official remote Raindrop MCP

```sh
rdrop connect raindrop
```

Endpoint: `https://api.raindrop.io/rest/v2/ai/mcp`. The [official documentation](https://developer.raindrop.io/mcp/mcp) currently describes a Pro beta service using Streamable HTTP and OAuth 2.1 or bearer auth. Configure auth in the client; generated output omits credentials. Local `rdrop` MCP instead uses REST and does not require that remote MCP service; individual Pro-only REST features still require Pro.

`RAINDROP_MCP_ENDPOINT` or `connect raindrop --url` overrides the config URL. It is not used by local `mcp serve` or the direct model loop.

## Executor

The source in [integrations/executor](../integrations/executor) is a portable app; its host calls Raindrop REST directly. It requires one Raindrop account whose API-key field is `token`. It offers cleanup counts, paged bookmarks, collection roots/children, tags, highlights, bookmark creation and tag append. It deliberately has a smaller surface than the local Go tools; bulk snapshots and context validation live in `rdrop`.

The app uses `never()` for reads and `always()` for its two writes. The host requires approval bound to the exact tool and arguments; `confirm:true` alone cannot bypass it. Requests use a fixed API origin and manual redirect handling, rejecting 3xx responses. Appending tags preserves existing names, with the same concurrent-edit limitations as any read-then-write operation.

The adapter is checked against `apps@0.0.1-beta.74`. Executor requires the **target host's exact framework version**, not `apps@latest`:

```sh
executor apps framework
npm ci --prefix integrations/executor
npm --prefix integrations/executor run check
npm --prefix integrations/executor test
node scripts/prepare-executor.mjs 0.0.1-beta.74
```

Proceed only if the host reports that exact version. A different version requires updating and verifying the adapter; the preparation script rejects a mismatch. Current [Executor authoring instructions](https://executor.sh/docs/author-an-app) explain the framework and provider account schema.

Create the app on your chosen local host after inspecting the prepared files:

```sh
executor apps create --name raindrop-cli --files dist/executor-app
executor apps source --app YOUR_APP_ID
executor apps deploy --app YOUR_APP_ID --commit EXACT_SOURCE_REVISION
```

Use the real source revision returned by the host. Bind your account through Executor's account UI; use its hosted options only for an intentionally selected hosted destination. The app has not been deployed by this repository's test suite.

For an agent to connect to Executor itself:

```sh
rdrop connect executor
rdrop connect executor --url https://mcp.executor.sh/org/YOUR_ORG/mcp
```

Local default: `http://127.0.0.1:4312/mcp`. Hosted connections use Executor's OAuth/PAT configuration in the client. Generating this config does not install, create, deploy or authenticate an Executor app. See [Executor MCP docs](https://executor.sh/docs/mcp).

## Other Raindrop MCP projects

[adeze/raindrop-mcp](https://github.com/adeze/raindrop-mcp) is a useful reference for dedicated MCP workflows. This project adds a shared shell/tool implementation, native packaging, exact snapshot plans, provider-neutral context and its own bounded direct-model loop. These are concrete differences, not a benchmark claim that one project's model accuracy or performance exceeds the other. Choose a transport and permission model that suits your host.
