# Release checklist

- [ ] Update `VERSION` and `CHANGELOG.md`.
- [ ] Confirm the supported firmware range against a physical Paper Pro.
- [ ] Review pinned XOVI/AppLoad releases and hashes.
- [ ] Run `scripts/build-release.ps1` from a clean clone.
- [ ] Confirm Go tests, vet, QML lint, ShellCheck, integration, CodeQL, and
      `govulncheck` pass in CI.
- [ ] Scan the staged Git tree and release ZIP for credentials/private keys.
- [ ] Verify the release ZIP against `SHA256SUMS.txt`.
- [ ] Install the exact ZIP payload on a clean/supported tablet.
- [ ] Test launch, real Device API refresh, offline cache, brightness restore,
      recovery, reinstall, uninstall, and reboot-to-stock behavior.
- [ ] Code-sign `TRMNL Installer.exe` when a trusted signing certificate is
      available; otherwise disclose the Windows SmartScreen warning.
- [ ] Create an annotated `vX.Y.Z` tag and push it.
- [ ] Review the automatically created draft GitHub Release, attach validation
      evidence, then publish it.
