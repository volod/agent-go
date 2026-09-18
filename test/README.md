# Test Layout

Unit and white-box tests sit beside their package as `*_test.go`; a test that needs unexported
identifiers stays there. This root-level `test/` directory follows the
[Go project-layout convention](https://github.com/golang-standards/project-layout/tree/master/test)
for everything else:

| Path | Purpose |
| --- | --- |
| `integration/` | Black-box tests that build `cmd/agent-go` and drive it as a process through its streams and exit codes |
| `testdata/` | Committed input data and golden outputs; Go ignores `testdata` directories as packages. Add it with its first file |

Rules:

- Tests are deterministic and network-free. Build fixtures in `t.TempDir()` or with
  `testing/fstest`; never commit large or binary fixtures when a test can generate them.
- Integration tests run in `go test ./...` and therefore in `make ci`. Put a slow or
  environment-dependent proof behind `//go:build integration` and add a Make target that runs it
  with `-tags integration`; `go vet` must also run with that tag.
- Tests that need a missing external tool skip with the reason; a declared run sets an
  environment variable that turns the skip into a failure.
- Reusable test helpers belong in a package under `test/` (for example `test/fixture/`), never in
  production packages.
