# Maintenance policy

The newest release is supported. Firmware compatibility is allowlisted, not
assumed. Security fixes take priority over features; a credential leak, unsafe
installer command, host-key bypass, archive traversal, unrecoverable device
change, or unbounded remote payload blocks release.

Maintainers should triage supported-version bugs within seven days when possible,
keep Dependabot enabled for Go and GitHub Actions, review CodeQL/govulncheck on
every release, and publish advisories for material vulnerabilities. Stale issues
may be closed after a documented request for reproduction evidence.

A release requires two independent evidence layers: automated host/CI validation
and installation of the exact final archive on supported physical hardware.
Compatibility claims must name model, firmware build, app version, and archive
SHA-256. See `RELEASE_CHECKLIST.md`.
