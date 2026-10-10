# Contributing

Build with the Go version in `go.mod`. Start behavior changes with focused package tests and finish with `make check`. Use fixtures for API and model transport; live checks need a disposable library and deliberate authorization.

```sh
make check
make check-integrations
```

Integration checks require Node 22.18+ (native TypeScript loading in adapter tests), npm and Python 3.10+. Run `npm ci --prefix integrations/executor` first. Laya adapter tests use a fake import and never download checkpoints. MCP tests use a real stdio client/server handshake. Model fixtures check the complete function-call/result round trip, bounds and rejection paths.

Keep the REST client, model transport and planning code in the standard library. The official MCP SDK and JSON Schema validator are justified protocol dependencies; avoid new dependencies that merely shorten a small helper. New API commands should follow stable official Raindrop endpoints and keep `raw` for JSON parity gaps.

The command modules in `cmd/rdrop` are shared by CLI dispatch and the structured catalog. Update `protocol_catalog.go` for tool exposure, and give writes explicit intent and appropriate annotations. Do not expose raw requests, arbitrary executable selection or local file reads through model tools. Host authorization is separate from `confirm:true`.

Preserve scriptable output: TSV for list defaults, JSON for structured data, bytes for transfers. Bulk changes need explicit scope, preview and visible apply. Test independent invariants such as stale rejection, all prechecks before writes, tag preservation, partial recovery and real protocol boundaries; avoid tests that just repeat the implementation.

Update help topics and run `make docs` after command changes. `make check` compares the generated command reference to the saved file. Update README, coverage, organizing/integration/model docs and changelog when behavior changes. Do not commit private contexts, plans, tokens, exports, model weights, node_modules or generated binaries.

CI covers Go checks and race tests, platform tests, optional adapters, native npm installation and container smoke checks. Check actual CI results before claiming a release is ready. [Packaging](docs/PACKAGING.md) describes deliberate tag-based GitHub releases and separate npm/container publication; a build does not authorize publishing a new namespace or hosted app.
