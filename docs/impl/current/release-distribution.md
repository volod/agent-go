# Release Distribution

## Packaging

`make dist` runs `make build-all`, clears `dist/` and runs `go run ./tools/dist`, which writes one
archive per `PLATFORMS` entry and `dist/SHA256SUMS`, as the
[specification](../../openspec/spec.md#release-distribution) requires.

- `tools/dist` reads each `bin/agent-go-<os>-<arch>[.exe]` with `debug/buildinfo` and names the
  archives after the stamped module version. Before writing any archive it refuses a missing
  binary, a malformed platform, a binary without a module version (`go run`, `-buildvcs=false`),
  binaries that report different versions, and, when `DIST_VERSION` is set, any other version.
- Linux and macOS archives are `.tar.gz`, Windows archives `.zip`. Each has one directory,
  `agent-go-<version>-<os>-<arch>/`, holding the executable (mode 0755) and `DIST_FILES`
  (`README.md`, `LICENSE`; mode 0644).
- Entry times are the commit time (`vcs.time`), or 1980-01-01 without one. Owners are 0, and the
  gzip header has no name or time, so rebuilding a commit with the same Go toolchain reproduces
  the checksums.
- `APP`, `PLATFORMS`, `DIST_DIR` and `DIST_FILES` live in `make/config.mk`; `make clean` removes
  `dist/`.

## Release workflow

`.github/workflows/release.yml` runs on pushed tags matching `v*.*.*`: `make ci`,
`make dist DIST_VERSION=<tag>`, `sha256sum -c SHA256SUMS`, then
`gh release create <tag> dist/* --verify-tag --generate-notes`. `ci.yml` runs `make dist` on
every push and pull request, so the release path is tested before a tag exists.

## Tests

`tools/dist/dist_test.go` covers the names, members, modes, times and checksums of both archive
formats, byte-identical output across runs, each refusal, commit-time parsing and the command's
exit codes.
