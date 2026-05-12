# Packaging

This project ships release archives from GitHub Actions. Extra package managers should be added only when they reduce real install friction.

## Current Release Assets

Tagged releases publish:

- macOS amd64 and arm64 tarballs
- Linux amd64 and arm64 tarballs
- Windows amd64 and arm64 zip files
- `checksums.txt`

The current `v0.1.0` macOS assets use Go's `darwin` target name. Future releases use friendlier `macos` artifact names.

## Dependabot

[`.github/dependabot.yml`](../.github/dependabot.yml) checks weekly for:

- GitHub Actions updates
- Go module updates

The project has no third-party Go dependencies today, but this keeps the setup ready if that changes.

## Homebrew

The repo includes a formula template at [`packaging/homebrew/rdrop.rb`](../packaging/homebrew/rdrop.rb).

This is not the same as being published in Homebrew core or a tap. To make Homebrew install work for users, create a tap repository such as:

```text
KingPsychopath/homebrew-tap
```

Then copy the formula to:

```text
Formula/rdrop.rb
```

Users could then install with:

```bash
brew tap KingPsychopath/tap
brew install rdrop
```

Before publishing a new tap formula release, update the `url` and `sha256` for the tagged source archive:

```bash
curl -L https://github.com/KingPsychopath/raindrop-cli/archive/refs/tags/v0.1.0.tar.gz | shasum -a 256
```

The formula intentionally builds from source with Go. That keeps it simple and avoids maintaining per-platform bottle metadata in this repository.
