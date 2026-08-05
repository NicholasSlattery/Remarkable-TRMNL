# Release checklist

## Source and automated gates

- [ ] Update `VERSION`, `CHANGELOG.md`, compatibility/validation docs, and links.
- [ ] Review pinned XOVI/AppLoad versions, upstream source archives, licenses,
      and SHA-256 values.
- [ ] Run `scripts/build-release.ps1` from a clean clone twice; confirm identical
      ZIP SHA-256 values.
- [ ] Confirm CI test and CodeQL jobs pass on the exact release commit: gofmt,
      race tests, vet, ShellCheck, Python compile, QML lint/resources, ARM64
      build, AppLoad protocol integration, `govulncheck`, and CodeQL.
- [ ] Inspect the staged tree and ZIP for credentials/private keys and review the
      CycloneDX SBOM plus third-party notices.
- [ ] Inspect `TRMNL Installer.exe` VersionInfo, icon/manifest, requested execution
      level, and Windows 10/11 launch behavior.
- [ ] Authenticode-sign and timestamp the installer. A publicly trusted
      organization/maintainer certificate is preferred; a self-signed community
      build must bundle its public certificate and clearly disclose that Windows
      will not trust it automatically.
- [ ] Verify the final ZIP against `SHA256SUMS.txt` and, for a public repository,
      verify the GitHub artifact attestation.

## Exact-release physical gates

- [ ] Back up/sync the test Paper Pro before Developer Mode or destructive tests.
- [ ] Confirm the supported firmware range against physical Paper Pro hardware.
- [ ] Install the **exact final ZIP payload** on a clean supported tablet.
- [ ] Test AppLoad launch, real Device API authentication/current/next, hosted
      plugin image, conditional refresh, rate limit behavior, and offline cache.
- [ ] Test color, fit/orientation, overlays/e-ink cleanup without extra API calls,
      frontlight restore on normal/crash exit, battery test, RTC scheduling,
      suspend/resume, Wi-Fi loss, and repeated launch/exit cycles.
- [ ] Test reboot-to-stock, **Reactivate after reboot**, stock restore, reinstall,
      uninstall-preserve, uninstall-and-erase, and official recovery guidance.
- [ ] Record model, OS build, app version, ZIP SHA-256, logs/captures, and every
      untested limitation in the release notes.

## Publication and advertising

- [x] Use a direct non-affiliate BYOD link with no tracking placeholder.
- [ ] Capture the real v2.0 screenshots/video listed in `LAUNCH.md`; remove private
      dashboard, fingerprint, device ID, IP, password, and API key data.
- [ ] Ensure README/release notes lead with Developer Mode factory-reset/security
      warnings and do not imply TRMNL/reMarkable endorsement.
- [ ] Create an annotated `vX.Y.Z` tag from the validated commit and push it.
- [ ] Review the automatically created draft release, generated notes, checksum,
      validation evidence, signature/provenance status, and known limitations.
- [ ] Publish the GitHub Release before linking advertisements to it.
- [ ] Enable private vulnerability reporting, branch protection/rulesets, required
      CI/CodeQL checks, and repository description/topics on the public repository.
