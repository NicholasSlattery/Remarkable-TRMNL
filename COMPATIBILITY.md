# Compatibility policy

| Component | v1.1 status |
|---|---|
| reMarkable Paper Pro (`Ferrari`) | Supported |
| reMarkable OS 3.26.x | Allowed by installer; exact-release physical check required |
| reMarkable OS 3.27.x | Supported; 3.27.3.0 previously exercised physically |
| Other reMarkable models | Blocked |
| OS 3.28+ or below 3.26 | Blocked until separately validated |
| XOVI 0.3.3 | Pinned |
| rm-xovi-extensions release 19 | Pinned |
| AppLoad 0.5.3 | Pinned |
| Windows 10/11 x64 | Graphical installer target |

Do not bypass installer model or firmware checks. reMarkable updates can change
private APIs, framebuffer behavior, runtime compatibility, or SSH access. Before
updating, return to the stock interface, check this document and open issues,
and keep a recovery path available.

New firmware support requires, at minimum: clean install, launch, Device API
refresh, cache/offline restart, touch/overlay cleanup, frontlight restore,
suspend/resume, reboot-to-stock, reactivation, uninstall/purge, and recovery
checks on physical hardware. Record the exact OS build and release ZIP hash.
