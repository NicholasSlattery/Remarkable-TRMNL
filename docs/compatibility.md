# Compatibility

| Component | Status |
|---|---|
| reMarkable Paper Pro (`Ferrari`) | Supported |
| reMarkable OS 3.27.x | Supported; 3.27.3.0 exercised on hardware |
| reMarkable OS 3.26.x | Allowed by the installer; not exercised on hardware |
| Other reMarkable models | Blocked |
| OS 3.28 and later, or below 3.26 | Blocked until validated |
| XOVI 0.3.3 | Pinned |
| rm-xovi-extensions release 19 | Pinned |
| AppLoad 0.5.3 | Pinned |
| Windows 10/11 x64 | Installer platform |

Do not bypass the model or firmware checks. reMarkable updates can change
private APIs, framebuffer behaviour, runtime compatibility, or SSH access.
Before updating the tablet, return to the stock interface, check this page and
the open issues, and keep a recovery path available.

Adding a firmware version to this list requires all of the following on physical
hardware, against the exact release archive: clean install, launch, a real
Device API refresh, cache and offline restart, overlay cleanup, front-light
restore, suspend and resume, reboot to stock, reactivation, uninstall, purge,
and stock recovery. Record the OS build and the archive's SHA-256.
