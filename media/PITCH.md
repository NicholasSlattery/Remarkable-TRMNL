# TRMNL for reMarkable Paper Pro — Full Pitch

**Version covered:** v2.0.0 · **Status:** release-ready pending exact-release physical
checks (`RELEASE_CHECKLIST.md`) · **License:** MIT (project code)

> **Required disclosure, everywhere this pitch is used:** Unofficial community
> software. Not affiliated with, endorsed by, or supported by TRMNL or reMarkable.
> Requires reMarkable Developer Mode, which factory-resets the tablet and reduces
> device security. Hosted TRMNL use requires a BYOD license.

---

## 1. The one-liner

Turn a Developer Mode reMarkable Paper Pro into a color TRMNL dashboard — installed
by a double-click Windows app, with a reversible return to the stock interface.

## 2. The elevator pitch (30 seconds)

The reMarkable Paper Pro is a 1620×2160 color e-paper tablet that people already
own and already love for one reason: it does one thing, quietly, without notifications.
TRMNL is a dashboard service built on exactly the same premise — a single calm
surface that shows you your day instead of competing for your attention.

The two have never talked to each other. This project makes them talk. It puts TRMNL
playlists on the Paper Pro's color display with scheduled refresh, an offline cache,
frontlight control, and on-device diagnostics — and it installs without a terminal,
without a soldering iron, and without a one-way door: every install action has a
matching recovery action.

## 3. The problem

Three problems stack on top of each other.

**For the reMarkable owner.** The Paper Pro is idle most of the day. It's a beautiful
color e-paper panel sitting face-down on a desk between writing sessions. There's no
first-party way to make it show anything but your own documents. The obvious use —
"show me my calendar, my tasks, the weather, my metrics, on the nicest screen I own" —
isn't available.

**For the TRMNL owner.** TRMNL's BYOD program explicitly invites you to bring your own
display. But "bring your own device" in practice has meant a Kindle jailbreak, a bare
e-paper panel wired to an ESP32, or an old Android tablet with the backlight bleeding
into a dark room. The nicest e-paper hardware most people already own — a reMarkable —
has had no supported path.

**For both.** The existing paths to custom software on a reMarkable are terminal-first.
They assume SSH, they assume you'll read a wiki, and they rarely tell you honestly what
you're giving up (Developer Mode is a factory reset and a security reduction) or how to
get back. That gap between "technically possible" and "safe for a normal person to try"
is where this project lives.

## 4. The solution

A native AppLoad application for the Paper Pro plus a local Windows installer that does
the dangerous parts carefully.

**On the tablet:** a QML frontend (`app/ui/TRMNL.qml`) over a static ARM64 Go backend
(`backend/cmd/trmnl-remarkable`) that speaks the TRMNL Device API, caches screens, and
manages refresh scheduling against the Paper Pro's real suspend cycle.

**On the desktop:** a single `TRMNL Installer.exe` that binds only to localhost, finds
the tablet over USB, shows you the SSH fingerprint before it trusts anything, validates
model, architecture, firmware, runtime checksums, free space, and payload paths, and
then installs transactionally with rollback.

**On the way out:** three distinct, clearly-labeled exits — restore the stock interface,
uninstall but keep settings, or uninstall and erase data — none of which quietly take
anything else with them.

## 5. What it actually does

### Display
- TRMNL playlists on the full 1620×2160 color e-paper panel
- Fit handling when a plugin sends a different aspect ratio
- Inversion toggle, and a native full-panel e-ink cleanup after every overlay
  transition (with a black/white fallback cycle when the native signal isn't available),
  so overlays never leave ghosts on the panel
- The last dashboard stays visible while new content loads — no "Waiting for dashboard"
  flash between screens

### Refresh and power
- Capability-detected RTC wake scheduling: the tablet wakes, refreshes, and goes back to
  sleep across normal Paper Pro suspend cycles, instead of staying awake to be useful
- The current image remains visible during normal suspend
- Honors the server's `Retry-After` instead of hammering a rate-limited endpoint
- Overlay navigation (opening controls, settings, diagnostics) does *not* consume a
  Device API request

### Offline
- Downloaded screens are cached locally and survive restarts and Wi-Fi loss
- Cached images are conditionally revalidated rather than blindly re-downloaded

### Controls
- Upper-right hotspot: refresh, next/previous screen, inversion, brightness, scheduled
  wake, history, diagnostics
- Complete frontlight controls in Settings
- Return to reMarkable three ways: the standard AppLoad swipe, an explicit button, or a
  two-second upper-left hold

### Diagnostics and the battery test
- On-device diagnostics with API keys redacted
- A persistent, no-terminal battery life test: charge, unplug, and it records 15-minute
  samples plus refresh and wake counts, survives suspend and app restarts, draws a trend
  chart, and projects runtime after at least a 2% drop

### Installation
- Double-click Windows installer, no terminal at any point
- Model, architecture, and firmware gates that should not be bypassed
- SSH host fingerprint shown for confirmation before trust
- Verified, checksum-pinned payloads and runtime components
- Post-reboot reactivation as a button, because a reboot intentionally starts the safest
  stock interface
- Windows icon, execution manifest, compatibility metadata, and version info; the release
  pipeline supports Authenticode signing

## 6. Why this is safe to try (the honest version)

This is the part most community projects skip, and it's the part that should sell it.

**What it costs you up front.** Developer Mode on a Paper Pro performs a factory reset,
enables root SSH, and reduces platform security. Cloud sync is not a backup for files
that haven't finished syncing. That cost is real, it's paid before this software touches
anything, and it is stated in the README's first warning block rather than in a footnote.

**What this app can and can't undo.** It can restore the stock interface, remove itself,
and erase its own data. It cannot turn Developer Mode back off — that requires
reMarkable's official software recovery, which can erase local data again. The project
says so plainly rather than implying reversibility it doesn't have.

**What it touches.** Only `/home/root`. It does not modify notebooks, documents, boot
partitions, or global power policy. Every persistent path is enumerated in `PRIVACY.md`.

**What leaves the device.** No maintainer-operated telemetry or analytics, at all. The
tablet talks directly to the TRMNL cloud or to your own BYOS origin. Device API requests
carry the device-scoped token, device ID/MAC, app version, model name, and battery
voltage — and nothing else.

**How credentials are handled.** The Device API key is stored in an owner-only `0600`
file, masked in the UI after save, never returned to QML, and redacted from diagnostics.
The tablet's SSH password lives in the localhost installer's browser memory for that
process only and is never saved or logged.

**How the network is constrained.** HTTPS required for production API and image traffic;
the only HTTP exception is a loopback development mock running on the tablet itself.
Credential-bearing redirects cannot change origin or protocol. Image bytes and dimensions
are bounded and validated. Server-sent firmware and reset directives are deliberately
ignored — the server can tell this client what to show, never what to become.

**How releases are verified.** CI blocks packaging on tests, `go vet`, QML lint,
ShellCheck, the ARM64 build, the AppLoad protocol harness, and `govulncheck`. Releases
ship a CycloneDX SBOM, SHA-256 checksums, and GitHub build-provenance attestations.

## 7. Who this is for

**Primary — the reMarkable owner who already thinks about attention.** They bought a
$500+ device that deliberately can't do email. They're the exact person who wants a
dashboard that shows and then shuts up.

**Secondary — the TRMNL BYOD owner shopping for a panel.** They've already bought into
the model and are choosing hardware. This makes a device they may already own into the
best-looking option on the list.

**Tertiary — the e-ink and self-hosting community.** BYOS support means the whole thing
runs against your own server with no hosted account, no BYOD license, and no third party
in the path.

**Explicitly not for:** anyone who needs warranty-safe hardware, anyone who keeps
sensitive material on the tablet, or anyone who wants a supported product with a support
line. And explicitly not for anyone who'd have to *buy* a Paper Pro to do this — if you
don't already own one, a dedicated dashboard is cheaper and better at the job. Say all of
this out loud in the listing; it filters out the people who'd be unhappy.

**On audience size, honestly.** The reachable group is Paper Pro owners × willing to
enable Developer Mode × wanting a dashboard. Independent analysis puts that in the
hundreds to low thousands worldwide, and nothing in the reMarkable or TRMNL communities
suggests otherwise. Plan accordingly: this is a well-made thing for a small group, not a
growth product. That's an argument for precision in the docs and against volume tactics
in the launch.

## 8. Why it's different

| | This project | Typical reMarkable hack | Typical BYOD panel |
|---|---|---|---|
| Install | Double-click Windows app | SSH + wiki + copy-paste | Solder / flash / wire |
| Display | 1620×2160 color e-paper you already own | — | Usually smaller, usually mono |
| Recovery | Three labeled exits, transactional rollback | "Reflash if it breaks" | N/A |
| Verification | Model/firmware/checksum/fingerprint gates | Trust the script | Trust the vendor |
| Honesty | Developer Mode cost stated first, in the README | Often buried | N/A |
| Telemetry | None | Varies | Varies |
| Self-host | Full BYOS, no hosted account needed | — | Varies |

The differentiator isn't "it works." It's that the risky parts are gated, the
irreversible parts are named, and the exits are built before the entrance.

## 9. Requirements

- reMarkable Paper Pro (`Ferrari`, ARM64) on reMarkable OS 3.26.x or 3.27.x
- Developer Mode enabled (factory reset — sync or export first)
- Windows 10/11 x64 and a USB cable, for the graphical installer
- One TRMNL BYOD license for hosted cloud use; **not** required for self-hosted BYOS
- The device-scoped **Device API key** from Developer Perks — not the `user_` account
  token

Other reMarkable models are blocked. OS 3.28+ and anything below 3.26 are blocked until
separately validated. See `COMPATIBILITY.md` before accepting a firmware update.

## 10. Known limitations — state these before anyone finds them

- **No automatic rotation.** "Auto" orientation follows image layout. The Paper Pro gives
  this app no reliable physical rotation signal.
- **Color depends on the plugin.** The panel displays color; the plugin or template must
  actually render it. Production auth and monochrome cloud output have been exercised;
  real-cloud color output must be rechecked for each release/plugin combination.
- **Reboot returns to stock by design.** Restarting the app after a reboot is a deliberate
  button press, not automatic.
- **`always_on` never overrides device power safety policy.** It's retained for
  compatibility and nothing more.
- **Firmware updates can break this.** It uses private interfaces on a device whose vendor
  never promised them.
- **Windows-only graphical installer** today.
- **Battery duration is not claimed.** The battery test measures *your* device; the
  project publishes no runtime number it hasn't collected.
- **Refresh rate is the real battery lever.** A purpose-built dashboard idles for months
  because its firmware does almost nothing between refreshes; a tablet running a full OS
  with Wi-Fi and suspend/resume cannot match that. Scheduled RTC wake recovers a lot of
  it, but expect the practical unplugged interval to be hours, not minutes. Start at a
  few refreshes per day and tighten from there while the battery test watches.

### Objections, answered

These four come up every time. Answer them before they're asked.

**"You've turned a $600 notebook into a dashboard — now you can't take notes."**
False, and it's the most common wrong assumption about this project. AppLoad runs
alongside the stock interface; TRMNL is one swipe away from being a notebook again
(center-top downward swipe, the **Return to reMarkable** button, or a two-second
upper-left hold), and a reboot lands on stock by design. The tablet keeps its day job.
The marginal cost of the dashboard is a BYOD license and the Developer Mode reset — not
the price of the device.

**"A dedicated dashboard costs a fraction of a Paper Pro."** True, and if you don't
already own a Paper Pro, buy the dedicated one — say so plainly. This is for hardware
that already exists on your desk and is idle most of the day. It is not an argument for
buying a reMarkable.

**"Firmware updates will break it."** They might. That's why `COMPATIBILITY.md` blocks
unvalidated firmware at the installer rather than letting you find out on the device,
why a reboot lands on stock, and why reactivation is a button. The failure mode is a
tablet that went back to being a tablet.

**"Isn't a dashboard just a well-mannered interruption?"** Fair hit. TRMNL and reMarkable
are not the same design philosophy — one pushes a scheduled sequence into your periphery,
the other waits for you to open it. Don't claim shared philosophy; claim shared screen
technology and idle hardware. The honest pitch is "your best e-paper panel is doing
nothing 22 hours a day," not "these two were made for each other."

## 11. Proof and current state

- v1.0.0 was physically verified on a Paper Pro running 3.27.3.0 across AppLoad launch,
  Device API behavior, cache, color display, controls, brightness restore, suspend/resume,
  Wi-Fi loss, crash safety, recovery, uninstall, and reboot safety.
- v2.0 validation is recorded in `V2.0_VALIDATION.md`.
- Exact-release physical checks remain required before publishing a release, and
  `RELEASE_CHECKLIST.md` gates the announcement on them.

Do not describe v2.0 as released until every blocking item in `RELEASE_CHECKLIST.md` is
complete and the draft GitHub Release is reviewed.

## 12. The ask

Depending on the audience, one of:

- **Users:** try it, and file a compatibility report with your exact OS build.
- **Contributors:** firmware validation on 3.26.x, a non-Windows installer path, and
  color-plugin verification are the three highest-value open areas.
- **TRMNL:** a listed community integration and clarity on BYOD custom-model color
  support would help users configure this correctly on the first try.
- **Everyone:** link the GitHub Release, not the executable, so people see checksums,
  requirements, and limitations before they see a download button.

## 13. Channel-ready copy

**Product listing name:** TRMNL for reMarkable Paper Pro

**One line:** Turn a Developer Mode reMarkable Paper Pro into a color TRMNL dashboard,
with a no-terminal Windows installer and reversible stock-interface recovery.

**Short description:** Display TRMNL playlists on the Paper Pro with color, scheduled
refresh, offline cache, frontlight controls, diagnostics, and a local Windows installer
that verifies the tablet and payload before installation.

**Release announcement:**

> TRMNL for reMarkable Paper Pro v2.0 brings TRMNL playlists to the 1620×2160 color
> e-paper display, including scheduled refresh, offline cache, frontlight controls,
> diagnostics, and a battery-life test. Installation is handled by a local Windows
> interface with model/firmware gates, SSH fingerprint confirmation, verified payloads,
> reactivation after reboot, recovery, and uninstall controls. Read the Developer Mode
> warning and verify the release checksum before use.

**Forum / community post opener:** "The Paper Pro is the nicest e-paper screen most of us
own, and it sits idle most of the day. I wanted TRMNL on it without a terminal and
without a one-way door, so I built both halves: the AppLoad app and a Windows installer
that checks the device before it touches it. Developer Mode is still a factory reset —
that's stated up front, and I'd rather lose the install than surprise someone."

**Topics / keywords:** `remarkable`, `remarkable-paper-pro`, `trmnl`, `epaper`, `eink`,
`dashboard`, `byod`, `qml`, `golang`, `windows-installer`

## 14. Asset shot list (real-device capture required)

Use actual v2.0 UI on exact-release hardware. Do not mock screenshots, and do not claim
real-cloud color, automatic rotation, signed-publisher status, firmware support, or
battery duration beyond collected evidence.

1. **Hero** — Paper Pro showing a colorful dashboard; no private calendar or personal data
2. **Installer** — preflight success, with fingerprint and password blurred
3. **Controls** — overlay, brightness, diagnostics, battery test
4. **Recovery** — the Reactivate / Restore / Uninstall actions side by side
5. **Video** — 20–40 seconds, uncut, install to dashboard

## 15. Things this pitch must never claim

- That it is affiliated with, endorsed by, or supported by TRMNL or reMarkable
- That Developer Mode is reversible from this app
- That the installer is signed, unless that release actually was
- Real-cloud color output, for any release/plugin pair that hasn't been rechecked
- Automatic rotation
- A specific battery runtime
- Firmware support outside 3.26.x–3.27.x
- Affiliate links without a commission disclosure — `BYOD_SETUP.md` intentionally
  uses a direct non-affiliate product link

## 16. FAQ

**Does this void my warranty?** Developer Mode may affect warranty and support for
problems it causes. That's reMarkable's call, not this project's.

**Can I undo Developer Mode?** Not from here. Only reMarkable's official software
recovery, which can erase local data again.

**Do I need to pay TRMNL?** For the hosted cloud, yes — one BYOD license for this device.
For a self-hosted BYOS server, no license and no hosted account.

**Will it survive a firmware update?** Unknown until validated. Return to the stock
interface and check `COMPATIBILITY.md` and open issues before updating.

**Does it touch my notebooks?** No. It writes only under `/home/root` and does not modify
documents, notebooks, boot partitions, or global power policy.

**Does it phone home?** No maintainer telemetry or analytics. The tablet talks only to the
TRMNL cloud or the BYOS origin you configure.

**Mac or Linux installer?** Not yet. The device-side scripts (`install.sh`,
`recover-stock.sh`, `uninstall.sh`) exist for advanced users; the graphical installer is
Windows 10/11 x64.

**Will it work on the reMarkable 2?** No. Other models are blocked by the installer.
