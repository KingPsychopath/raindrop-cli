# Get started with rdrop for Raindrop.io

Install `rdrop`, connect your Raindrop account, then use your shell or agent. An existing Codex/Claude agent already supplies the model: you do not need another AI account, Laya, Executor or ACP to get started.

## 1. Install

Choose one:

```sh
brew install KingPsychopath/tap/rdrop
# Or Node 22+, with optional dependencies enabled:
npm install --global @machineyearn/raindrop-cli@0.2.1
rdrop version
```

Without a global npm install, replace `rdrop` in the commands below with `npx --yes @machineyearn/raindrop-cli@0.2.1`. [Native downloads, source builds and checksums](PACKAGING.md).

## 2. Connect Raindrop

Create a token in [Raindrop's integration settings](https://app.raindrop.io/settings/integrations). On macOS/Linux, prepare a private configuration file:

```sh
mkdir -p "$HOME/.config/rdrop"
chmod 700 "$HOME/.config/rdrop"
touch "$HOME/.config/rdrop/env"
chmod 600 "$HOME/.config/rdrop/env"
```

Open that file in your text editor and add `RAINDROP_TOKEN=your_token_here`, replacing the placeholder privately. On Windows, the default file is `%USERPROFILE%\.config\rdrop\env`; keep it accessible only to your account. Then:

```sh
rdrop doctor
rdrop cleanup report --json
```

The doctor checks configuration and account access without printing the token. The report reads your library; it does not change bookmarks. Environment variables override configuration files. If you use `XDG_CONFIG_HOME` or another location, run `rdrop config paths` to see the paths this installation reads. [Configuration reference](COMMANDS.md#config).

## 3. Use your agent or shell

Copy the [setup and organizing prompt from the README](../README.md#3-ask-your-agent) into your agent. It tells the agent to check setup, use existing tools, start with a report and small sample, then show exact changes before applying an approved plan.

For Codex or Claude, install the [shared plugin](CLIENTS.md#codex-and-claude-plugins). Other local clients can use the [stdio MCP configuration](CLIENTS.md#direct-stdio-configuration). The default MCP server reads and plans. Installing the plugin does not enable writes.

Or start directly:

```sh
rdrop collections --tree
rdrop search 'type:article' --json
rdrop cleanup report --json
```

Your existing tag names and collection paths are the starting vocabulary. Review a bounded sample before extending organization to the rest of the library. [Reviewed planning and recovery](ORGANIZING.md).

## Add capabilities when you need them

| Need | Use |
| --- | --- |
| Raindrop's semantic search, extracted content or AI audits | Add `--official` to `mcp serve` or `agent`; [upstream plan limits apply](OFFICIAL_MCP.md) |
| Run a tool-capable model without an external agent | [`rdrop agent`](MODELS.md#direct-tool-calling-loop) with a hosted or local model |
| Fixed categories with visible uncertainty | [`classify`](MODELS.md#typed-decision-providers) with configured Jev, Laya or Clef |
| Local model inference | [Qwen/Laya setup and verification](VERIFICATION.md#start-the-actual-local-models) |
| Portable tools in Executor | [Executor setup](INTEGRATIONS.md#executor) |
| Give this MCP server to an existing ACP agent | [ACP configuration](INTEGRATIONS.md#acp) |

These are optional routes. Keep the first setup small, then choose a provider or host for a concrete job.

## If setup fails

- **Command not found:** restart the terminal, check the npm/Homebrew bin directory is on PATH, or use `npx`. GUI agents may need an absolute executable path.
- **Missing token or HTTP 401:** inspect `rdrop config paths` and `rdrop config show`, then replace the token in the private file. Do not paste it into chat.
- **Missing npm platform package:** reinstall with `npm install --global --include=optional @machineyearn/raindrop-cli@0.2.1`, or use a native archive.
- **No agent tools:** reload/start a new client session and check its MCP connection. Browser-only clients require a hosted HTTP service; this local server uses stdio.
- **Official/model provider unavailable:** check that provider's configuration and account limits. The shell CLI and normal REST reads do not depend on that provider.
