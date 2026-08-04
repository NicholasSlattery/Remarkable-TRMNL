# Changelog

## 1.1.0 - 2026-08-04

- Redraw and refresh the dashboard after opening or closing controls, settings,
  diagnostics, and after resume so e-ink overlays cannot remain stuck.
- Request a native full-panel e-ink cleanup after every overlay transition and
  every five seconds while a menu is open, with a black/white fallback cycle.
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
