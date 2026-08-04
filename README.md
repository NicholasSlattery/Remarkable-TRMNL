# TRMNL for reMarkable Paper Pro

TRMNL dashboards on the reMarkable Paper Pro, with color, scheduling, offline
cache, frontlight controls, diagnostics, and a safe return to the stock
interface.

> **Unofficial software:** This project is not affiliated with, endorsed by, or
> supported by TRMNL or reMarkable. It uses community extension runtimes and
> root/SSH access. A firmware update can make it incompatible.

## Install without a terminal

1. Download `TRMNL-for-reMarkable-1.1.0-Windows-x64.zip` from Releases.
2. Extract the ZIP and double-click **TRMNL Installer.exe**.
3. Connect the Paper Pro by USB, enable SSH/developer access in the tablet's
   settings, and paste the password shown by the tablet.
4. Click **Find my tablet**, confirm the displayed fingerprint, and click
   **Install TRMNL**.
5. On the tablet, open **AppLoad**, then tap **TRMNL**.

The installer checks the model, architecture, firmware, runtime checksums, free
space, and device fingerprint before changing anything. It also has graphical
**Restore stock interface** and **Uninstall TRMNL** buttons. See
[`EASY_INSTALL.md`](EASY_INSTALL.md) for detailed steps and troubleshooting.

### Supported configuration

- reMarkable Paper Pro (`Ferrari`, ARM64)
- reMarkable OS 3.26.x or 3.27.x
- Windows 10/11 x64 for the graphical installer
- XOVI 0.3.3, extension release 19, and AppLoad 0.5.3

Do not force installation on another model or firmware version. Open an issue
with the diagnostics output so compatibility can be tested safely first.

## First run

Open the small upper-right hotspot in TRMNL, choose **Settings**, and enter the
Device API key from your TRMNL account. Custom BYOS servers must use HTTPS;
unencrypted HTTP is accepted only for a loopback development mock on the tablet.
The API key is masked, stored in an owner-only `0600` file, omitted from logs,
and never returned to the frontend after saving.

Use the upper-right hotspot for refresh, next/previous screen, inversion,
brightness, scheduled wake, history, and diagnostics. The dashboard redraws and
checks for an update whenever controls or settings are opened or closed, so an
overlay cannot remain ghosted on the e-ink panel. Return to reMarkable with the
normal AppLoad center-top downward swipe, the **Return to reMarkable** button, or a
two-second hold in the upper-left corner.

## Important behavior and limitations

- A reboot intentionally returns to the safest stock interface. Open the
  graphical installer and click Install again, or run `/home/root/xovi/start`
  if you are an advanced user, to reactivate AppLoad.
- “Auto” orientation follows content layout; the Paper Pro does not expose a
  reliable physical auto-rotation signal to this application.
- **Sleep between updates and wake for refresh** is enabled by default. When the
  firmware exposes a writable RTC wake alarm, TRMNL programs it for the next
  server-provided refresh interval. The e-ink image remains visible while the
  tablet uses its normal suspend behavior; on resume TRMNL refreshes immediately.
  Settings reports clearly when scheduled wake is unavailable.
- `always_on` is retained for configuration compatibility. TRMNL does not disable
  the tablet's safety-critical global power policy or force Linux to suspend.
- Production TRMNL authentication and monochrome cloud output were tested.
  Full-color rendering was tested on the physical panel with the compatible
  mock server; an active real-cloud color plugin should be rechecked for every
  release.

## Safety and recovery

TRMNL stores files only under `/home/root`. It does not touch notebooks,
documents, boot partitions, or global power settings. Uninstall removes only the
TRMNL application and preserves the shared XOVI/AppLoad runtime so other
extensions cannot be deleted accidentally.

Persistent state:

- Settings: `/home/root/.config/trmnl-remarkable/config.json`
- Cache: `/home/root/.cache/trmnl-remarkable/`
- Logs/history: `/home/root/.local/share/trmnl-remarkable/`
- App bundle: `/home/root/xovi/exthome/appload/trmnl-remarkable/`

If the graphical installer cannot connect, rebooting the tablet removes the
temporary XOVI injection and starts the stock interface.

## Build from source

Requirements: Go 1.25+, Qt 6 `rcc`, PowerShell 5.1+, and internet access for
verified upstream runtime downloads.

```powershell
./scripts/build.ps1
./scripts/build-release.ps1
```

`build.ps1` runs formatting checks, unit tests, `go vet`, ARM64
cross-compilation, QML resource compilation, and bundle validation.
`build-release.ps1` downloads pinned upstream assets, verifies SHA-256 hashes,
creates a normalized device archive, builds the graphical installer, generates
an SBOM and checksums, and writes the final ZIP under `release/`.

## Architecture

- `app/ui/TRMNL.qml`: AppLoad QML frontend.
- `backend/cmd/trmnl-remarkable`: static ARM64 backend for the Device API,
  scheduling, cache, settings, history, battery, and frontlight.
- `installer`: local-only graphical host installer using verified SSH.
- `install.sh`, `recover-stock.sh`, `uninstall.sh`: bounded device operations.

The client uses `GET /api/display` for playlist advance and tries
`GET /api/display/current`, then `GET /api/current_screen`, for a current-screen
refresh. Firmware/reset directives from the API are deliberately ignored.

## Security, support, and licensing

Report vulnerabilities privately using [`SECURITY.md`](SECURITY.md). General
help is covered in [`SUPPORT.md`](SUPPORT.md); contribution checks are in
[`CONTRIBUTING.md`](CONTRIBUTING.md).

Project code is MIT licensed. The easy installer redistributes unmodified,
checksum-pinned XOVI/AppLoad runtime components under their respective
LGPL-3.0/GPL-3.0 licenses. Full notices and corresponding source locations are
in [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).
