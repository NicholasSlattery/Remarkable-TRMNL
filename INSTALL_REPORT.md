# Installation report

Status: **Version 1.0.0 release payload installed and verified on the connected Paper Pro**

## Detected environment

- Device: `reMarkable Ferrari` (reMarkable Paper Pro)
- OS: reMarkable/Codex Linux `3.27.3.0` (`5.7.126` userspace)
- Architecture: `aarch64`; kernel `6.12.49+git-imx8mm-ferrari-g68b95e858a0a`
- App framebuffer: 1620 x 2160, RGBA, 6528-byte stride; portrait panel
- Xochitl: Qt 6.8.2, active/enabled; executable SHA-256
  `9749880daa2f10844e77b560ec0ecddd1634d43eb328af637c7026edf3ef120e`
- XOVI: 0.3.3; AppLoad: 0.5.3; extension release: 19.0.0
- Compatibility gate: AppLoad requires reMarkable OS `>=3.26,<3.28`; 3.27.3.0 passed
- Root filesystem: read-only; all persistent project files are under `/home/root`
- Free space before install: 46,941,712 KiB under `/home/root`
- Frontlight: `/sys/class/backlight/rm_frontlight`, range 0..2047,
  read/write/read verified; original test value 1734

## Installed architecture and paths

- Xochitl-hosted AppLoad QML frontend with a static ARM64 Go backend
- App bundle: `/home/root/xovi/exthome/appload/trmnl-remarkable`
- Source/tools/report: `/home/root/trmnl-remarkable`
- Settings: `/home/root/.config/trmnl-remarkable/config.json` (atomic, mode 0600)
- Cache: `/home/root/.cache/trmnl-remarkable`
- Logs/history/installation state: `/home/root/.local/share/trmnl-remarkable`
- XOVI/AppLoad runtime: `/home/root/xovi`; shims: `/home/root/shims`
- TRMNL protocol: `GET /api/display` plus current-screen fallbacks,
  `access-token` and optional `ID` headers, server refresh/timeout fields

## Physical tests performed and passed

- AppLoad icon visible from the stock sidebar; fullscreen launch and clean backend handshake
- First-run settings UI with masked key field and no credential required
- Local TRMNL-compatible API authentication, response fetch, PNG download, atomic
  cache, full-color display, 1200x1600 portrait and 1800x1000 landscape images
- Production TRMNL cloud authentication with the supplied Device API Key;
  `/api/display` returned a screen, the tablet downloaded and displayed it, and a
  manual refresh succeeded with an atomic cached copy
- TRMNL account device profile changed to Custom Device at 1620x2160, PNG, large
  framework, portrait, and 16.7-million-color presentation; the live response was
  verified as an exact 1620x2160 PNG
- Server-provided 60-second schedule, manual refresh, next/current protocol paths,
  duplicate cache reuse, prior cached screen, and refresh history
- Endpoint loss and HTTP 401 kept the cached screen visible with useful errors;
  manual recovery succeeded when the endpoint returned
- Offline restart displayed cache and retained bounded retry behavior
- Controls open/close, diagnostics, redacted key, inversion, battery, and status
- Frontlight slider physically verified at 0/512/1024/1536/2047 for
  0/25/50/75/100%; normal exit and `SIGKILL` both restored 1734
- Return button, two-second upper-left hold, and AppLoad center-top downward swipe/X
- Three consecutive launch/exit cycles; no orphan backend; Xochitl remained active
- Deep suspend from 07:51:52 to 07:52:06 UTC; same Xochitl/backend PIDs after resume
- Wi-Fi interface down/up; cached display and backend survived; interface restored
- Idle 30-second sample: backend RSS 9,760 KiB and 0 CPU ticks; Xochitl 2 CPU
  ticks; cache 88 KiB; no tight loop
- Recovery disabled the bundle, unmounted the XOVI drop-in, restarted vendor
  Xochitl, and verified no XOVI mapping or preload environment remained
- Default uninstall removed the project source and bundle, preserved the shared
  XOVI/AppLoad runtime, and returned to active stock Xochitl; clean reinstall passed
- Tablet reboot changed boot ID and returned to active stock Xochitl with no
  XOVI injection; AppLoad was then reactivated and the clean first-run UI relaunched
- Host: Go unit tests/vet, ARM64 bundle self-check, QML resource compilation/lint,
  and Linux `SOCK_SEQPACKET` AppLoad/mock integration
- Release: the exact normalized 1.0.0 device payload from the graphical-installer
  ZIP passed checksum verification, transactional install, device self-check,
  file-mode checks, XOVI injection, AppLoad journal detection, and local/remote
  binary SHA-256 comparison on 2026-08-04

## Security and integrity

- XOVI release SHA-256: `32d64d1262ddc984e3235c7d0340a398fe6d5b3efa6a979865f5977b32630d27`
- AppLoad release and Vellum SHA-512 values matched; installed `appload.so`
  SHA-256: `31214cbbe64c8bfe7d99096f077c3009dba8a42ef1a733801aa0ec59c134e7cc`
- API keys are never returned to QML after save, are redacted from diagnostics,
  and are not logged. Final installed state contains no mock key.
- No notebooks, documents, boot partitions, or global power policies were changed.

## Recovery and uninstall

- Stock recovery: `/home/root/trmnl-remarkable/recover-stock.sh`
- Uninstall (preserve settings/cache and shared extension runtime): `/home/root/trmnl-remarkable/uninstall.sh`
- Full purge: `/home/root/trmnl-remarkable/uninstall.sh --purge`

## Genuine limitations

- The production playlist currently has no active content and therefore rendered
  TRMNL's monochrome "You're caught up" screen. The account and client are configured
  for full color, and Paper Pro color rendering was physically verified with the
  on-device compatible mock; a real cloud color plugin was not present to re-test.
- XOVI uses a RAM-backed systemd drop-in. A reboot intentionally returns to safest
  stock mode; run `/home/root/xovi/start` once after a reboot to restore the AppLoad
  sidebar entry. It is not auto-started because a persistent preload would weaken
  recovery safety on this read-only-root firmware.
- Paper Pro does not expose an application auto-rotation signal here. Portrait and
  landscape content and the manual orientation setting work; automatic physical
  rotation was not independently available to test.
- “Always on” is stored for compatibility but deliberately does not alter the
  tablet's global suspend policy.
