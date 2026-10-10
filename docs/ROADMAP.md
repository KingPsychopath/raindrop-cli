# Current scope and remaining work

## Current scope (v0.2.1)

- Native Go CLI with stable REST commands, tree paths, sanitized TSV and JSON/byte pipelines.
- Local MCP stdio server using the official Go SDK, read-only defaults, validated inputs, organizing prompt/resource and explicit mutation exposure.
- ACP session configuration for supplying tools to an existing ACP agent; remote Raindrop and Executor MCP configurations.
- Bounded direct-model loop for Responses and compatible Chat Completions providers, with exact plan capture.
- Existing-tag/collection context, explicit naming policy, provider-neutral multi-tag/move proposals and account-bound snapshot apply/recovery.
- Optional System One HTTP decisions (Jev/Laya-compatible), trusted local command decisions and a Laya Python adapter.
- Cloudflare Clef decision adapter, searchable tool schemas, and shared Codex/Claude plugin packaging.
- Optional official MCP reads for semantic search, extracted content and AI-assisted library audits, alongside reviewed bulk creation.
- Executor portable REST app with real exact-call host approval, checked against an exact framework version.
- Six native release targets, npm platform packages/launcher under `@machineyearn`, checksums and a non-root container.
- Command reference generated from CLI help; source Go version used in CI.
- Public identity **rdrop for Raindrop.io**, stable technical names, shared agent plugins, and npm-backed MCP Registry discovery with verified release bytes and short-lived publishing identity.

## Runtime evidence

- Real Laya HTTP and command inference and a real Qwen tool loop produced an exact plan, passed dry-run and preserved existing tags and notes when applied to a disposable fixture.
- The local Executor deployment has a connected account and passed its provider health check and live cleanup/tag/collection reads.
- [Verification commands and limits](VERIFICATION.md) record the tested model/runtime versions. [Packaging](PACKAGING.md) describes the exact release artifacts and publication process.

## Evaluation still needed

- Evaluate Jev, Laya and other local models on reviewed bookmark examples; fixture checks do not prove semantic quality, performance or equal calibration.
- Jev live inference requires an API key. Representative live write journeys require an intentionally selected disposable Raindrop library; no live bookmark mutation was performed during this release verification.
- Hosted Clef live inference requires a Workers AI token/account. Local Clef serving and additional note sources such as Obsidian need separate runtime/storage adapters and evidence.

## Future options, with separate evidence

Persistent sessions and streaming; resumable full-library scans with rate-aware read retries; a public HTTP/OAuth MCP boundary; native ACP agent implementation only if a real host requires it; finer tool exposure profiles; provider evaluation datasets and local model lifecycle tooling.

Do not implement unattended organization merely because a tool loop exists. Default remains report, bounded context, visible proposal, reviewed exact plan, intentional apply. Mutations need uncertain-outcome recovery before retries or scheduling become more automatic.
