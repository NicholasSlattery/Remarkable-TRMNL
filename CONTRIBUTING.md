# Contributing

By contributing, you agree that your contribution is licensed under the MIT
license used by this project.

## Development checks

Install Go 1.25+ and Qt 6 (`rcc` and `qmllint` on `PATH`), then run:

```powershell
gofmt -w backend installer
go test ./...
go vet ./...
./scripts/build.ps1
./scripts/test-linux-integration.ps1
```

Do not add `_secrets`, `_downloads`, `_tools`, `_upstream`, device captures,
credentials, or generated release artifacts. Device-changing pull requests must
describe rollback behavior and demonstrate recovery on supported hardware.

Version changes require `VERSION`, `CHANGELOG.md`, compatibility evidence, the
exact final-archive smoke test, and updated upstream hashes if runtime versions
change.

Pull requests must use the template and distinguish host/mock evidence from
physical-device evidence. Do not weaken model/firmware, HTTPS, host-key, archive,
image-size, or recovery gates merely to make an unsupported setup install.
