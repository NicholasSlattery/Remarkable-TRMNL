<div align="center">

# TRMNL for reMarkable Paper Pro

**Turn a reMarkable Paper Pro into a full-color, battery-aware TRMNL dashboard.**

[![Latest release](https://img.shields.io/github/v/release/NicholasSlattery/Remarkable-TRMNL?display_name=tag&sort=semver)](https://github.com/NicholasSlattery/Remarkable-TRMNL/releases/latest)
[![CI](https://github.com/NicholasSlattery/Remarkable-TRMNL/actions/workflows/ci.yml/badge.svg)](https://github.com/NicholasSlattery/Remarkable-TRMNL/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-222222.svg)](LICENSE)
[![Device: Paper Pro](https://img.shields.io/badge/device-reMarkable%20Paper%20Pro-8a6d3b.svg)](COMPATIBILITY.md)

</div>

<p align="center">
  <img src="docs/images/remarkable-trmnl-demo.jpeg" alt="TRMNL dashboard running on a reMarkable Paper Pro" width="100%">
  <br>
  <em>A TRMNL dashboard running on a reMarkable Paper Pro in landscape.</em>
</p>

TRMNL for reMarkable adds scheduled dashboard refreshes, an offline cache,
front-light controls, battery testing, history, diagnostics, and safe recovery
to the stock reMarkable interface. Normal installation and removal use a guided
Windows app—no terminal required.

**[Install](#quick-install)** · **[Connect TRMNL](#connect-to-trmnl)** ·
**[Save battery](#battery-friendly-settings)** · **[Troubleshoot](#troubleshooting)** ·
**[Read the docs](#documentation)**

> [!CAUTION]
> **Developer Mode factory-resets the Paper Pro.** Sync or export every
> important document before enabling it. Developer Mode reduces device security
> and may affect support for problems caused by modifications. Uninstalling this
> app does not turn Developer Mode off; leaving it requires reMarkable's official
> [software recovery](https://support.remarkable.com/s/article/Software-recovery)
> and may erase local data again. Read reMarkable's
> [Developer Mode guidance](https://support.remarkable.com/s/article/Developer-mode)
> before continuing.

This is unofficial community software. It is not affiliated with, endorsed by,
or supported by TRMNL or reMarkable. Firmware updates can break compatibility.

## At a glance

| | |
|---|---|
| **Current release** | [v2.0.0](https://github.com/NicholasSlattery/Remarkable-TRMNL/releases/tag/v2.0.0) |
| **Supported tablet** | reMarkable Paper Pro (`Ferrari`, ARM64) |
| **Supported firmware** | reMarkable OS 3.26.x and 3.27.x |
| **Installer computer** | Windows 10 or 11, x64 |
| **Recommended connection** | Direct USB cable; Wi-Fi is also supported |
| **TRMNL cloud** | One claimed BYOD license and its Device API key |
| **Self-hosted TRMNL** | HTTPS BYOS server; no hosted BYOD license required |
| **Normal install experience** | Graphical installer; no command line |

The installer blocks unsupported models, architectures, and firmware instead of
guessing. Check [COMPATIBILITY.md](COMPATIBILITY.md) before updating the tablet.

## Why use it?

| Feature | What it gives you |
|---|---|
| **Color-capable dashboard** | Displays color when the selected TRMNL plugin and device model render color. |
| **Battery-aware scheduling** | Uses normal suspend and RTC wake alarms when the firmware supports them. |
| **Offline resilience** | Keeps the last good dashboard cached and visible through network interruptions. |
| **On-device controls** | Refresh, next/previous screen, inversion, orientation, fit, front light, history, and diagnostics. |
| **Battery-life test** | Records 15-minute samples and estimates runtime after a valid 10% discharge window. |
| **Recovery-first design** | Reboots to the stock interface and provides one-click reactivation, restore, uninstall, and purge actions. |
| **Security guardrails** | Verifies the device, host fingerprint, payload, runtime checksums, firmware, architecture, and free space. |

## Quick install

### 1. Prepare the tablet

- Confirm it is a **reMarkable Paper Pro** on **3.26.x or 3.27.x**.
- Sync or export everything you care about.
- Enable **Developer Mode** and complete the required reset/onboarding flow.
- Display the tablet's current SSH password.
- Charge above 20%, wake and unlock the tablet, then connect it directly by USB.
- If using TRMNL cloud, complete the short [BYOD setup](BYOD_SETUP.md).

### 2. Download and verify v2.0.0

Download both files from the
[v2.0.0 release](https://github.com/NicholasSlattery/Remarkable-TRMNL/releases/tag/v2.0.0):

- `TRMNL-for-reMarkable-2.0.0-Windows-x64.zip`
- `SHA256SUMS.txt`

In PowerShell, verify that the displayed hash exactly matches the value in
`SHA256SUMS.txt`:

```powershell
Get-FileHash .\TRMNL-for-reMarkable-2.0.0-Windows-x64.zip -Algorithm SHA256
Get-Content .\SHA256SUMS.txt
```

Extract the **entire ZIP**. Do not run the installer from inside the archive.

> [!NOTE]
> The v2.0 community installer is Authenticode-signed with a self-signed
> certificate and timestamped. Windows can still show **Unknown publisher** or
> a SmartScreen warning because the certificate is not publicly trusted. Verify
> the checksum and review [SELF_SIGNED_RELEASE.md](SELF_SIGNED_RELEASE.md) before
> deciding whether to run it. The ZIP contains the public certificate only—not
> its private key.

### 3. Run the installer

1. Double-click **TRMNL Installer.exe** with the `payload` folder beside it.
2. Leave `10.11.99.1` selected for a USB connection.
3. Paste the SSH password currently shown by the tablet.
4. Click **Find my tablet**.
5. Confirm the model, firmware, and displayed SSH fingerprint.
6. Click **Install TRMNL** and keep the cable connected until it reports success.
7. On the tablet, open **AppLoad**, then tap **TRMNL**.

The password stays only in browser memory for the local installer session. It
is not written to disk or logs, and the installer listens only on `127.0.0.1`.
For expanded instructions, see [EASY_INSTALL.md](EASY_INSTALL.md).

## Connect to TRMNL

Open the small hotspot in the upper-right corner of the tablet screen and choose
**Settings**.

### TRMNL cloud

1. Open the claimed BYOD device on TRMNL.
2. Copy its **Device API Key** from **Developer Perks**.
3. Paste that key into the tablet app.
4. Confirm the pre-filled Device ID matches the BYOD device's Wi-Fi MAC.
5. Tap **Test connection**, then **Save**.

Do not use the Account API token that begins with `user_`; it is for a different
API. The full walkthrough is in [BYOD_SETUP.md](BYOD_SETUP.md).

### Self-hosted BYOS

Choose the custom server option and enter its device identity and **HTTPS**
origin. Plain HTTP is accepted only for a loopback development mock running on
the tablet.

## Everyday controls

| Action | Where to find it |
|---|---|
| Open controls or Settings | Tap the upper-right hotspot |
| Refresh or change screens | Use the refresh and previous/next buttons |
| Adjust the front light | Use the brightness control or follow system brightness |
| View refresh history | Open **History** |
| Export useful status | Open **Diagnostics**; secrets are redacted |
| Return to reMarkable | Use the AppLoad center-top downward swipe, the on-screen button, or hold the upper-left corner for two seconds |

Opening or closing an overlay performs a local e-ink cleanup. It does not use a
TRMNL Device API request.

## Battery-friendly settings

A static e-ink image remains visible without keeping the app awake. Battery life
is driven mainly by wake frequency, Wi-Fi activity, and the front light.

| Setting | Recommended value |
|---|---|
| reMarkable **Auto-sleep** | **On** |
| reMarkable **Light sleep** | **On** |
| reMarkable **Auto power-off** | **Off** for scheduled dashboard wakeups |
| TRMNL **Wake for refresh** | **On** |
| Front light | Off or as low as practical |
| Refresh interval | Longer intervals for substantially better runtime |

The current dashboard remains visible during normal suspend. When supported,
the app schedules an RTC alarm, lets the tablet sleep, wakes for the refresh,
and returns to sleep.

For a realistic estimate, charge the tablet, unplug it, start the battery test,
and let it discharge by at least 10%. The estimate is intentionally invalidated
if charging is detected or the battery percentage increases.

## Recovery and removal

A reboot intentionally returns to the safest stock reMarkable interface. Run
the installer again, find and verify the tablet, then choose the action you need:

| Installer action | Result |
|---|---|
| **Reactivate after reboot** | Starts XOVI/AppLoad without reinstalling files. |
| **Restore stock interface** | Stops the temporary runtime injection but keeps app data. |
| **Uninstall TRMNL** | Removes the app while preserving its settings/cache and shared XOVI/AppLoad runtime. |
| **Uninstall and erase data** | Also removes TRMNL settings, cache, history, logs, and battery-test records. |

None of these actions disables Developer Mode. Use reMarkable's official
[software recovery](https://support.remarkable.com/s/article/Software-recovery)
to leave Developer Mode.

TRMNL writes only under `/home/root`. It does not modify notebooks, documents,
boot partitions, or the tablet's global power policy. See [PRIVACY.md](PRIVACY.md)
for the exact persistent paths.

## Compatibility

| Component | v2.0 status |
|---|---|
| reMarkable Paper Pro (`Ferrari`) | Supported |
| reMarkable OS 3.26.x | Installer-allowed; review validation notes |
| reMarkable OS 3.27.x | Supported; 3.27.3.0 has been exercised physically |
| Other reMarkable models | Blocked |
| OS 3.28+ or below 3.26 | Blocked until separately validated |
| Windows 10/11 x64 | Supported graphical installer platform |

Do not bypass model or firmware checks. Before installing a reMarkable update,
return to the stock interface and review [COMPATIBILITY.md](COMPATIBILITY.md).

## Troubleshooting

| Problem | What to try |
|---|---|
| Tablet is not found | Wake and unlock it, reconnect directly without a hub, confirm Developer Mode, and copy the current SSH password again. |
| USB address does not connect | Use `10.11.99.1`; for Wi-Fi, enter the tablet's current local IP instead. |
| SSH fingerprint changed | Do not bypass it. Reconnect by USB and independently verify the tablet before accepting a new fingerprint. |
| HTTP 401 or 403 | Confirm the BYOD device is claimed and use its Device API Key—not a `user_...` Account API token. |
| HTTP 429 | Wait for the server's retry window; the app honors `Retry-After`. |
| Dashboard is monochrome | Confirm the TRMNL model/plugin is configured to render full color; the tablet cannot add color absent from the source image. |
| App is gone after reboot | Use **Reactivate after reboot** in the installer; this stock-first behavior is intentional. |
| Battery drains quickly | Enable Auto-sleep and Light sleep, disable Auto power-off, lower the front light, and increase the refresh interval. |
| Windows warns about the publisher | Verify the ZIP checksum and read the bundled self-signing disclosure before proceeding. |

Still stuck? Read [SUPPORT.md](SUPPORT.md) before opening an issue, and never
include an API key, SSH password, private document, or unredacted diagnostic.

## Security and privacy

- The graphical installer binds only to localhost and validates the device,
  firmware, architecture, free space, SSH fingerprint, payload paths, and
  checksums before modifying the tablet.
- The Device API key is masked, stored in an owner-only `0600` file, omitted
  from logs, and never returned to QML after saving.
- Remote servers and image downloads require HTTPS. Credential-bearing
  cross-origin or protocol-changing redirects are refused.
- Cached images are conditionally revalidated and bounded by byte and dimension
  limits before decoding.
- Production code ignores server firmware/reset directives and never sends the
  tablet's SSH password to TRMNL.

Please report security issues using the private process in
[SECURITY.md](SECURITY.md), not a public issue.

## Documentation

| Topic | Guide |
|---|---|
| Guided installation | [EASY_INSTALL.md](EASY_INSTALL.md) |
| TRMNL cloud/BYOD setup | [BYOD_SETUP.md](BYOD_SETUP.md) |
| Supported devices and firmware | [COMPATIBILITY.md](COMPATIBILITY.md) |
| Windows self-signing and trust | [SELF_SIGNED_RELEASE.md](SELF_SIGNED_RELEASE.md) |
| Privacy and stored data | [PRIVACY.md](PRIVACY.md) |
| Security policy | [SECURITY.md](SECURITY.md) |
| Support and diagnostics | [SUPPORT.md](SUPPORT.md) |
| Release changes | [CHANGELOG.md](CHANGELOG.md) |
| v2.0 validation record | [V2.0_VALIDATION.md](V2.0_VALIDATION.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Third-party components | [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) |

<details>
<summary><strong>Build and test from source</strong></summary>

### Requirements

- Go 1.25.12 or newer
- Qt 6.8.2 with `rcc` and `qmllint`
- PowerShell 5.1 or newer
- Python 3
- ShellCheck for CI-equivalent validation
- Internet access for checksum-pinned runtime downloads

### Commands

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
ARM64 builds, the AppLoad protocol harness, and `govulncheck`. Public archives
also receive GitHub build-provenance attestations.

</details>

<details>
<summary><strong>Project architecture</strong></summary>

- `app/ui/TRMNL.qml` — AppLoad QML frontend
- `backend/cmd/trmnl-remarkable` — static ARM64 Device API and scheduling backend
- `installer` — localhost graphical installer using verified SSH
- `install.sh`, `recover-stock.sh`, `uninstall.sh` — bounded device operations

The Device API client advances with `GET /api/display` and uses compatible
current-screen endpoints for refreshes. It sends the TRMNL `access-token`,
optional device ID, app version, and real battery voltage.

</details>

## Contributing and license

Issues and pull requests are welcome. Start with
[CONTRIBUTING.md](CONTRIBUTING.md) and the
[Code of Conduct](CODE_OF_CONDUCT.md).

Project code is available under the [MIT License](LICENSE). The installer
redistributes checksum-pinned XOVI/AppLoad components under their respective
LGPL/GPL licenses; details and source locations are in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
