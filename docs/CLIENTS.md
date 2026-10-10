# Connect your agent

Use the native `rdrop` executable or the pinned npm launcher from a local MCP client. The server starts read-only, reads the normal private CLI configuration, and keeps stdout reserved for MCP messages. v0.2.0 npm publication is currently waiting for the owner's required 2FA setup; use native stdio configuration below until the [publishing status](PACKAGING.md#npm-packages) confirms registry availability.

## Codex and Claude plugins

The shared package in [integrations/agent-plugin](../integrations/agent-plugin) includes a portable Agent Plugins manifest, Codex/Claude compatibility manifests, a read-only MCP connection and one organizing skill. The skill teaches tool discovery, existing-tag preservation, bounded context, uncertainty review and exact-plan apply. It does not enable mutation tools or configure credentials.

For Codex:

```sh
codex plugin marketplace add KingPsychopath/raindrop-cli --ref v0.2.0
codex plugin add raindrop@machineyearn
```

Supported desktop versions also expose this marketplace in their plugin directory. Reload/start a new session after installation. Older clients can use the direct MCP configuration below. See [OpenAI's plugin packaging guide](https://developers.openai.com/plugins/build/plugins).

For Claude Code, run these inside Claude:

```text
/plugin marketplace add KingPsychopath/raindrop-cli
/plugin install raindrop@machineyearn
```

Invoke `/raindrop:organize` or ask it to organize your Raindrop library. For local development, use `claude --plugin-dir ./integrations/agent-plugin`. The manifests and marketplace were checked with `claude plugin validate`; this is package validation, not proof of every client's UI workflow. See [Claude's plugin reference](https://code.claude.com/docs/en/plugins-reference).

The MCP command is pinned to `@machineyearn/raindrop-cli@0.2.0`. Updating the plugin does not automatically authorize writes. Keep Node 22+ and optional npm dependencies enabled. Configure the token through the private CLI env/token file and run `rdrop doctor` before starting the client.

## Direct stdio configuration

For clients using `mcpServers`, including Claude Desktop and Cursor, add:

```json
{
  "mcpServers": {
    "raindrop": {
      "command": "npx",
      "args": ["--yes", "@machineyearn/raindrop-cli@0.2.0", "mcp", "serve"]
    }
  }
}
```

Or `rdrop connect mcp --command /absolute/path/rdrop` prints native configuration. GUI apps need an executable they can find; use an absolute command path when their PATH differs from your terminal. Leave tokens out of project configuration. A client running inside a container/remote workspace uses that environment's private token configuration.

| Client | Connection |
| --- | --- |
| Codex | Plugin above, or a local stdio MCP server |
| Claude Code / Claude Desktop | Plugin or local stdio server |
| Cursor / Windsurf | Add the stdio server in that client's MCP settings |
| VS Code / Copilot | Use its `servers` configuration with `type: "stdio"` and the same command/args |
| Zed / other local MCP hosts | Use their native stdio configuration format with the same command/args |
| Claude web/mobile, ChatGPT web and other remote-only hosts | Require a reachable HTTP MCP service; the local stdio command cannot serve this role |

These are transport-compatible setup routes, not a claim that every listed client has been independently tested here or that Raindrop lists this project as an official integration. See the [official supported-client page](https://help.raindrop.io/integrations/mcp#supported-clients) for Raindrop's hosted HTTP/OAuth service.

## Optional hosted reads and other connections

Add `--official` to the local server's arguments to include [Raindrop's official AI read tools](OFFICIAL_MCP.md). The client remains local stdio; the server makes an authenticated outbound connection to the official hosted service. Free/Pro limitations still apply.

`rdrop connect acp` supplies the local MCP server to an existing ACP agent. `connect raindrop` prints HTTP configuration for the official service; `connect executor` prints configuration for your Executor host. These are different connections with their own auth. [Integration details](INTEGRATIONS.md).
