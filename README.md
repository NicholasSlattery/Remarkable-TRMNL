# TRMNL for reMarkable Paper Pro

TRMNL dashboards on a reMarkable Paper Pro, with color, scheduling, an offline
cache, frontlight controls, diagnostics, and a safe return to the stock
interface.

> [!WARNING]
> This app requires reMarkable **Developer Mode**. Enabling Developer Mode on a
> Paper Pro performs a **factory reset**, reduces the device's security, and may
> affect warranty/support for problems caused by Developer Mode. Sync or export
> every important document before enabling it. Uninstalling this app does not
> disable Developer Mode; leaving it requires reMarkable's official software
> recovery process. Read [reMarkable's Developer Mode guidance](https://support.remarkable.com/s/article/Developer-mode)
> before continuing.

This is unofficial community software. It is not affiliated with, endorsed by,
or supported by TRMNL or reMarkable. A firmware update can break compatibility.

## What you need

- A reMarkable Paper Pro (`Ferrari`, ARM64) on reMarkable OS 3.26.x or 3.27.x.
- A Windows 10/11 x64 computer and a USB cable for the graphical installer.
- Developer Mode enabled and the SSH password displayed by the tablet.
- For the hosted TRMNL cloud, one TRMNL BYOD license for this device. A
  self-hosted BYOS server does not require the hosted BYOD license.

See [BYOD_SETUP.md](BYOD_SETUP.md) for purchasing/claiming BYOD, obtaining the
correct Device API key, and configuring the Paper Pro model. See
[COMPATIBILITY.md](COMPATIBILITY.md) before accepting a firmware update.

## Install without a terminal

1. Download `TRMNL-for-reMarkable-2.0.0-Windows-x64.zip` and
   `SHA256SUMS.txt` from the same GitHub Release.
2. Verify the ZIP checksum, extract the whole ZIP, and double-click
   **TRMNL Installer.exe**.
3. Connect the Paper Pro by USB and paste the SSH password shown by the tablet.
4. Click **Find my tablet**, verify the model/firmware and displayed SSH
   fingerprint, then click **Install TRMNL**.
5. On the tablet, open **AppLoad**, tap **TRMNL**, then open Settings and enter
   the TRMNL Device API key.

The installer binds only to localhost and validates the model, architecture,
firmware, runtime checksums, free space, payload paths, and device fingerprint
before changing the tablet. Detailed steps are in [EASY_INSTALL.md](EASY_INSTALL.md).

Windows releases include an application icon, requested-execution manifest,
Windows 10/11 compatibility metadata, and file/product version information.
Authenticode signing is supported by the release pipeline. The v2.0 community
build is self-signed, so Windows can still report an unknown or untrusted
publisher until its included public certificate is trusted locally. Always
verify the ZIP SHA-256 before deciding whether to run it.

## Developer Mode, clearly

Developer Mode is a device-level reMarkable setting, not this app's settings
screen. On supported Paper Pro firmware, open **Settings > General > Software**,
tap the software version, open **Advanced**, and choose **Developer Mode**.
Follow the on-device warning and setup flow; menu wording may vary.

- Enabling it erases local tablet data. Cloud sync is not a backup for files
  that have not finished syncing, so confirm sync or export first.
- It enables root SSH and reduces normal platform security. Do not keep
  sensitive material on a modified device unless you accept that risk.
- This installer can restore the stock interface and remove its own files, but
  neither action turns Developer Mode off.
- To leave Developer Mode, use reMarkable's official
  [software recovery](https://support.remarkable.com/s/article/Software-recovery),
  which can erase local data again.

## First run and controls

Open the small upper-right hotspot in TRMNL and choose **Settings**. Enter the
Device API key from the TRMNL device's **Developer Perks** area. Do not enter the
Account API key that begins with `user_`. The app pre-fills the tablet's detected
Wi-Fi MAC as Device ID; confirm it matches the BYOD device in TRMNL.

Custom BYOS servers must use HTTPS. Plain HTTP is accepted only for a loopback
development mock running on the tablet. The API key is masked, stored in an
owner-only `0600` file, omitted from logs, and never returned to QML after save.

The upper-right controls provide refresh, next/previous screen, inversion,
brightness, scheduled wake, history, and diagnostics. Opening or closing an
overlay performs a local e-ink cleanup only; it does not consume a Device API
request. Return to reMarkable with the normal AppLoad center-top downward swipe,
the **Return to reMarkable** button, or a two-second upper-left hold.

Settings also includes a battery-life test. After charging and unplugging, it
records lightweight 15-minute samples, refreshes, and wake events, survives
normal suspend/app restarts, rejects tests that include charging, and estimates
runtime after at least a 10% drop.

## Reboot, recovery, and removal

A reboot intentionally starts the safest stock interface. To start AppLoad
again, run the graphical installer, find/verify the tablet, and click
**Reactivate after reboot**. Advanced users may run `/home/root/xovi/start`.

The installer offers three distinct actions:

- **Restore stock interface** stops the temporary runtime injection without
  deleting app data or disabling Developer Mode.
- **Uninstall TRMNL** removes the app and preserves its settings/cache plus the
  shared XOVI/AppLoad runtime used by other extensions.
- **Uninstall and erase data** additionally purges TRMNL settings, cache,
  history, logs, and battery-test records. It still does not disable Developer
  Mode or remove shared runtimes.

TRMNL writes only under `/home/root`; it does not modify notebooks, documents,
boot partitions, or global power policy. Persistent paths are documented in
[PRIVACY.md](PRIVACY.md).

## Behavior and limitations

- "Auto" orientation follows image layout; this app has no reliable physical
  rotation signal from the Paper Pro.
- Wake-for-refresh is enabled by default when firmware exposes a writable RTC
  alarm. For dashboard use, enable reMarkable Auto-sleep and Light sleep, leave
  Auto power-off disabled, and use a longer minimum refresh interval. The
  current image remains visible during normal suspend.
- `always_on` is retained for compatibility but never overrides the tablet's
  safety-critical power policy.
- Production authentication and monochrome cloud output have been exercised.
  Real-cloud color output must be rechecked for each release/plugin combination.
- v2.0 validation is recorded in [V2.0_VALIDATION.md](V2.0_VALIDATION.md).
  Exact-release physical checks remain required before publishing a release.

## Build and test from source

Requirements: Go 1.25+, Qt 6 `rcc`/`qmllint`, PowerShell 5.1+, Python 3,
ShellCheck (for the CI-equivalent checks), and internet access for pinned
upstream runtime downloads.

```powershell
./scripts/build.ps1
./scripts/test-linux-integration.ps1
./scripts/build-release.ps1 -Version 2.0.0
```

`build.ps1` runs formatting, unit tests, `go vet`, ARM64 cross-compilation,
QML resource compilation, and bundle validation. `build-release.ps1` verifies
pinned runtime hashes, creates normalized archives, embeds Windows resources,
builds the graphical installer, generates a CycloneDX SBOM and checksums, and
writes the release ZIP under `release/`.

The release workflow blocks packaging on tests, vet, QML lint, ShellCheck,
ARM64 build, the AppLoad protocol harness, and `govulncheck`. Public releases
also receive GitHub build-provenance attestations. Maintainers can configure
`WINDOWS_SIGNING_CERT_BASE64` and `WINDOWS_SIGNING_CERT_PASSWORD` secrets to
Authenticode-sign the installer before packaging.
Maintainers without a publicly trusted certificate can instead run
`scripts/sign-self-signed-release.ps1`; this signs the executable and bundles
the public certificate, but does not remove Windows trust warnings.

## Architecture and API policy

- `app/ui/TRMNL.qml`: AppLoad QML frontend.
- `backend/cmd/trmnl-remarkable`: static ARM64 Device API/scheduling backend.
- `installer`: localhost graphical installer using verified SSH.
- `install.sh`, `recover-stock.sh`, `uninstall.sh`: bounded device operations.

The client uses `GET /api/display` to advance and tries
`GET /api/display/current`, then `GET /api/current_screen`, for current-screen
refresh compatibility. It sends the TRMNL `access-token`, optional device `ID`,
app firmware version, and real battery voltage. It honors `Retry-After`, refuses
cross-origin or protocol-changing redirects, requires HTTPS for image downloads,
conditionally revalidates cached images, limits image bytes/dimensions, and
deliberately ignores server firmware/reset directives.

## Security, support, and licensing

Read [SECURITY.md](SECURITY.md), [PRIVACY.md](PRIVACY.md),
[SUPPORT.md](SUPPORT.md), and [CONTRIBUTING.md](CONTRIBUTING.md).

Project code is MIT licensed. The installer redistributes checksum-pinned
XOVI/AppLoad components under their respective LGPL/GPL licenses. Notices,
versions, hashes, and corresponding-source locations are in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
