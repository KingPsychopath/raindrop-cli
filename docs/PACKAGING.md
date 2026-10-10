# Builds and distribution

`rdrop` is a native Go executable. npm is an optional launcher distribution for clients that prefer `npx`; it does not duplicate the Go command or tool implementation. Docker is another packaging option. Archive checksums, per-binary npm integrity and container digests identify different artifacts and are complementary.

## Native source and releases

```sh
make check
make install
./bin/rdrop version
```

The Go version is declared in `go.mod`; CI uses that file. The release workflow builds macOS, Linux and Windows binaries for amd64 and arm64, with `CGO_ENABLED=0`, trimpath and version/commit/date metadata. GitHub release archives include README and license; `checksums.txt` covers downloadable archives and npm tarballs. Build timestamps mean that separate invocations are not promised to be byte-for-byte reproducible.

For a released archive, download it and the matching `checksums.txt` from the same [GitHub release](https://github.com/KingPsychopath/raindrop-cli/releases). On Linux run `sha256sum --check checksums.txt --ignore-missing`; on macOS compare `shasum -a 256 YOUR_ARCHIVE` to that archive's entry before extracting it. Do not reuse another release's checksums.

The supplied [Homebrew formula](../packaging/homebrew/rdrop.rb) retains its existing v0.1.0 source URL and verified checksum. Update both when publishing a new release; its `head` option tracks main. A formula's description does not imply that an old tagged version contains development features.

## npm packages

```sh
node --test packaging/npm/launcher.test.cjs
node scripts/build-npm.mjs 0.2.0-development.1
node scripts/pack-npm.mjs
```

Outputs under `dist/npm`:

- `@machineyearn/raindrop-cli`: Node launcher, exact optional platform dependencies.
- `@machineyearn/rdrop-{darwin,linux,windows}-{amd64,arm64}`: native executable, integrity metadata and license.

The root source manifest is private to prevent accidental publication. Generated manifests are versioned and publishable. Override the scope with `RDROP_NPM_SCOPE` while building if publishing under another owned scope. The unscoped `raindrop-cli` name is already occupied by another package. The authenticated publisher `machineyearningengineer` is the verified owner of the `machineyearn` npm organization.

The launcher requires Node 22+, selects the installed native package, verifies its SHA-256 before execution and passes stdio/exit status through. There are no install scripts or runtime downloads. Install with optional dependencies enabled. Integrity detects a changed executable relative to its packaged metadata; it does not authenticate a compromised publisher.

Authenticate with `npm login`. Publish the exact tarballs downloaded from the matching GitHub release, so npm and the release checksums refer to the same bytes:

```sh
gh release download v0.2.0 --dir dist --pattern '*.tgz' --pattern checksums.txt
# Verify every downloaded tarball against checksums.txt first.
node scripts/publish-npm.mjs 0.2.0
```

The script checks all seven package manifests, publishes each native package before the launcher, and skips an existing version only when its registry SHA-512 equals the local tarball. It stops on authentication, policy, network or integrity errors; it never overwrites a published version. npm may require browser/2FA approval. `RDROP_RELEASE_DIR` selects another artifact directory and `RDROP_NPM_SCOPE` another owned scope. The release workflow attaches `.tgz` files to GitHub; it does **not** publish npm packages or provision credentials. Install a published release with:

```sh
npx --yes @machineyearn/raindrop-cli@0.2.0 mcp serve
```

Use a pinned real release version and treat it as installed executable code. The launcher and all six platform packages share the same release version.

## Docker and digests

```sh
docker build --build-arg VERSION=dev --build-arg COMMIT="$(git rev-parse HEAD)" -t rdrop:local .
docker run --rm rdrop:local version
docker run --rm -i --env RAINDROP_TOKEN rdrop:local mcp serve
```

The image contains a statically linked executable and CA certificates in a non-root distroless runtime. `-i` carries MCP stdio; omit `-t` so terminal framing does not corrupt the protocol. Passing an env name inherits the value from the shell without writing it in the command. Docker does not automatically inherit the user's config files; mount a private token file and set `RAINDROP_TOKEN_FILE` if preferred.

This project supplies a Dockerfile, without an automatically published registry image. After intentionally pushing an image, record the registry manifest digest and use `IMAGE@sha256:...` in production. `docker image inspect rdrop:local --format '{{.Id}}'` identifies a local image; it is not a registry manifest digest. The Dockerfile pins verified base manifest digests as well as readable tags. Refresh those digests deliberately with image verification.

## Release procedure

Run `make check` and `make check-integrations`, inspect the diff and release notes, and choose a new semantic version. Only then push its `vX.Y.Z` tag. CI builds six native archives and seven npm tarballs, verifies local npm installation on its host, builds/smokes Docker in CI, and publishes GitHub assets with checksums. It does not deploy Executor, download Laya checkpoints, mutate Raindrop data or publish npm/container packages.
