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

The project has a Homebrew tap:

```bash
brew tap KingPsychopath/tap
brew install rdrop
```

Tap repository:

```text
https://github.com/KingPsychopath/homebrew-tap
```

The source repo keeps a copy of the formula at [`packaging/homebrew/rdrop.rb`](../packaging/homebrew/rdrop.rb). The live formula is maintained in the tap at `Formula/rdrop.rb`.

For a new release:

1. Update the tap formula `url` to the new tag.
2. Update the tap formula `sha256`.
3. Test the formula locally.
4. Commit and push the tap update.

Get the source archive checksum with:

```bash
curl -L https://github.com/KingPsychopath/raindrop-cli/archive/refs/tags/v0.1.0.tar.gz | shasum -a 256
```

The formula intentionally builds from source with Go. That keeps it simple and avoids maintaining bottle metadata for now.
