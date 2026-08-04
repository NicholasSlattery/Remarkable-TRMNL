# Changelog

## 1.1.0 - 2026-08-04

- Correct Device API metadata to send battery voltage, honor `Retry-After`,
  securely revalidate stable image URLs, and prevent overlay navigation from
  consuming API calls.
- Refuse credential-bearing cross-origin/protocol redirects and non-loopback
  HTTP image URLs; bound and validate downloaded image bytes and dimensions.
- Add graphical post-reboot reactivation, explicit data purge, clearer
  Developer Mode/reset/recovery warnings, and TRMNL BYOD onboarding.
- Add Windows icon/manifest/version metadata, optional Authenticode signing,
  public build provenance, gated release validation, and community launch files.
- Redraw and refresh the dashboard after opening or closing controls, settings,
  diagnostics, and after resume so e-ink overlays cannot remain stuck.
- Request a native full-panel e-ink cleanup after every overlay transition,
  with a black/white fallback cycle when the native signal is unavailable.
- Keep the last dashboard visible while refreshed content loads, avoiding a
  repeated “Waiting for dashboard” flash in Settings.
- Add a persistent, no-terminal battery life test with 15-minute sampling,
  sleep/wake and refresh counts, a trend chart, and projected runtime.
- Add complete front-light controls to the Settings page.
- Add capability-detected RTC wake scheduling for battery-friendly refreshes
  across normal Paper Pro suspend cycles.

All notable changes are documented here. This project follows semantic
versioning.

## 1.0.0 - 2026-08-04

- Initial public-ready Paper Pro release.
- Added a double-click graphical Windows installer with device discovery,
  fingerprint confirmation, payload verification, recovery, and uninstall.
- Added HTTPS-only remote BYOS validation and loopback-only HTTP development.
- Added transactional runtime rollback and shared-runtime-safe uninstall.
- Added reproducible device packaging, release checksums, SBOM, CI, and public
  security/support/contribution documentation.
- Physically verified AppLoad launch, Device API behavior, cache, color display,
  controls, brightness restore, suspend/resume, Wi-Fi loss, crash safety,
  recovery, uninstall, and reboot safety on a Paper Pro running 3.27.3.0.
