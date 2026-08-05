# Launch and advertising kit

Do not advertise v2.0 as released until every blocking item in
`RELEASE_CHECKLIST.md` is complete and the draft GitHub Release is reviewed.

## Product listing copy

**Name:** TRMNL for reMarkable Paper Pro

**One line:** Turn a Developer Mode reMarkable Paper Pro into a color TRMNL
dashboard, with a no-terminal Windows installer and reversible stock-interface
recovery.

**Short description:** Display TRMNL playlists on the Paper Pro with color,
scheduled refresh, offline cache, frontlight controls, diagnostics, and a local
Windows installer that verifies the tablet and payload before installation.

**Required disclosure:** Unofficial community software. Not affiliated with
TRMNL or reMarkable. Requires reMarkable Developer Mode, which factory-resets
the tablet and reduces security. Hosted TRMNL use requires a BYOD license.

## Suggested release announcement

> TRMNL for reMarkable Paper Pro v2.0 brings TRMNL playlists to the 1620x2160
> color e-paper display, including scheduled refresh, offline cache, frontlight
> controls, diagnostics, and a battery-life test. Installation is handled by a
> local Windows interface with model/firmware gates, SSH fingerprint confirmation,
> verified payloads, reactivation after reboot, recovery, and uninstall controls.
> Read the Developer Mode warning and verify the release checksum before use.

Link the announcement to the GitHub Release first, not directly to an executable,
so users see checksums, validation evidence, requirements, and known limitations.

## Launch assets still requiring real-device capture

- Hero image: Paper Pro showing a colorful dashboard; no private calendar/data.
- Installer image: preflight success with fingerprint and password blurred.
- Controls image: overlay, brightness, diagnostics, and battery test.
- Recovery image: Reactivate/Restore/Uninstall actions.
- 20-40 second uncut install-to-dashboard video.

Use actual v2.0 UI and exact-release hardware. Do not mock screenshots or claim
real-cloud color, automatic rotation, signed-publisher status, firmware support,
or battery duration beyond collected evidence.

Recommended topics/keywords: `remarkable`, `remarkable-paper-pro`, `trmnl`,
`epaper`, `eink`, `dashboard`, `byod`, `qml`, `golang`, `windows-installer`.

## BYOD link disclosure

`BYOD_SETUP.md` uses a direct, non-affiliate product link. Never imply TRMNL
endorsement or add affiliate tracking without a clear commission disclosure.
