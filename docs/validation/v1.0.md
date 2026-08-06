# Acceptance verification checklist

> This table records the earlier physical validation campaign (primarily v1.0
> on Paper Pro firmware 3.27.3.0). See `V2.0_VALIDATION.md` for current release
> evidence. Rows that were not repeated against the exact v2.0 archive remain
> historical regression context rather than fresh v2.0 completion claims.

`PASS` is backed by host tests, device state, framebuffer captures, sysfs
read-back, service journals, or process measurements from the connected tablet.

| Requirement | Status | Evidence |
|---|---|---|
| Native AppLoad icon/fullscreen launch | PASS | Sidebar/AppLoad/framebuffer capture and AppLoad journal |
| Paper Pro ARM64 and OS compatibility | PASS | Ferrari, aarch64, OS 3.27.3.0; AppLoad compatibility gate |
| TRMNL/BYOS API and authentication | PASS (mock) | On-device mock success, 401 negative test, Linux protocol integration |
| Color, portrait, landscape, aspect-fit | PASS | 1620x2160 RGBA captures; 1200x1600 and 1800x1000 cached test images |
| Manual refresh, next/current, schedule | PASS | Device history and protocol integration; 60-second mock interval |
| Secure settings/cache/history | PASS | 0600 config, atomic tests, redacted diagnostics, offline restart |
| Controls, inversion, battery, diagnostics | PASS | Physical touch and framebuffer captures |
| Brightness 0/25/50/75/100 and restore | PASS | sysfs values 0/512/1024/1536/2047; normal/crash restore to 1734 |
| Return, fallback hold, AppLoad swipe | PASS | Physical synthetic-touch tests and AppLoad termination journals |
| Repeated cycles/crash safety | PASS | Three cycles, forced crash, Xochitl active, no orphan backend |
| Suspend/resume and Wi-Fi loss | PASS | 14-second deep suspend; same PIDs; wlan0 down/up; recovery |
| Idle CPU/memory | PASS | Backend 9,760 KiB RSS, 0 ticks/30s; Xochitl 2 ticks/30s |
| Stock recovery | PASS | XOVI mount/mapping absent and vendor Xochitl active after recovery |
| Reboot stock safety | PASS | New boot ID; Xochitl active with no preload; AppLoad later relaunched |
| Real TRMNL account | PASS | Production Device API authentication, screen download, and manual refresh on the physical tablet |
| Real cloud full-color plugin | NOT RUN | Production account had no active color plugin; physical full-color display was verified with the compatible mock |
| Automatic physical rotation | NOT AVAILABLE | No Paper Pro application rotation signal exposed; manual mode tested |
