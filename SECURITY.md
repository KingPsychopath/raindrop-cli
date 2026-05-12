# Security

## Reporting

Please report security issues privately to the maintainer before opening a public issue.

## Token Handling

`raindrop-cli` reads `RAINDROP_TOKEN` from the environment or from `~/.config/openclaw/gateway.env`. The token is never intentionally written by this tool.

Do not paste tokens into issues, pull requests, examples, screenshots, or shell history. Prefer environment variables or your shell's secret manager integration.

## Scope

This CLI sends authenticated requests to the official Raindrop.io API. Review bulk commands carefully before running them against a large library, especially delete, tag merge, collection merge, and empty-trash operations.
