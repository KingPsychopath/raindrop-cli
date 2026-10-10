# Current scope and remaining work

## Implemented in development source

- Native Go CLI with stable REST commands, tree paths, sanitized TSV and JSON/byte pipelines.
- Local MCP stdio server using the official Go SDK, read-only defaults, validated inputs, organizing prompt/resource and explicit mutation exposure.
- ACP session configuration for supplying tools to an existing ACP agent; remote Raindrop and Executor MCP configurations.
- Bounded direct-model loop for Responses and compatible Chat Completions providers, with exact plan capture.
- Existing-tag/collection context, explicit naming policy, provider-neutral multi-tag/move proposals and account-bound snapshot apply/recovery.
- Optional System One HTTP decisions (Jev/Laya-compatible), trusted local command decisions and a Laya Python adapter.
- Executor portable REST app with real exact-call host approval, checked against an exact framework version.
- Six native release targets, buildable npm platform packages/launcher, checksums and a non-root container.
- Command reference generated from CLI help; source Go version used in CI.

## Release and runtime evidence still needed

- Publish a reviewed GitHub release, update the Homebrew tag/checksum and verify downloads.
- Verify npm namespace ownership and intentionally publish platform packages followed by the launcher. No npm release is implied by build artifacts.
- Deploy the Executor adapter to an explicitly selected compatible host and bind an account.
- Evaluate Jev, Laya and other local models on reviewed bookmark examples; fixture checks do not prove semantic quality, performance or equal calibration.
- Verify representative live API journeys on a disposable library. Unit/integration fixtures never mutate the user's account.

## Future options, with separate evidence

Persistent sessions and streaming; resumable full-library scans with rate-aware read retries; a public HTTP/OAuth MCP boundary; native ACP agent implementation only if a real host requires it; finer tool exposure profiles; provider evaluation datasets and local model lifecycle tooling.

Do not implement unattended organization merely because a tool loop exists. Default remains report, bounded context, visible proposal, reviewed exact plan, intentional apply. Mutations need uncertain-outcome recovery before retries or scheduling become more automatic.
