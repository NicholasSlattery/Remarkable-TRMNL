# Easy installation guide

No command line is required.

## Before you begin

> **Developer Mode erases the tablet.** Confirm every document is synced or
> exported before enabling it. Developer Mode reduces security and issues caused
> by it may not be covered by reMarkable warranty/support. Uninstalling TRMNL
> does not disable Developer Mode; leaving it requires official software recovery.

1. Confirm this is a reMarkable Paper Pro on firmware 3.26.x or 3.27.x.
2. Read reMarkable's [Developer Mode article](https://support.remarkable.com/s/article/Developer-mode).
3. Sync/export your documents, enable Developer Mode, complete the reset and
   onboarding, then display the tablet's SSH password.
4. Charge above 20% and connect the tablet directly to Windows by USB.
5. If you use TRMNL cloud, complete the [TRMNL setup](trmnl-setup.md) first.

## Verify the download

Download the ZIP and `SHA256SUMS.txt` from the same GitHub Release. In PowerShell:

```powershell
Get-FileHash .\TRMNL-for-reMarkable-2.1.0-Windows-x64.zip -Algorithm SHA256
```

The value must exactly match `SHA256SUMS.txt`. Extract the entire ZIP; do not run
the installer from inside the archive.

## Install

1. Double-click **TRMNL Installer.exe** with the `payload` folder beside it.
2. Leave `10.11.99.1` for USB and paste the current SSH password.
3. Click **Find my tablet**.
4. Confirm **reMarkable Ferrari**, supported firmware, and the SSH key. The
   installer remembers the key and tells you on later runs whether it matches.
5. Click **Install TRMNL** and leave the cable connected until success.
6. On the tablet, open **AppLoad**, tap **TRMNL**, then configure the Device API
   key in Settings.

The password exists only in browser memory for this local installer session. It
is not written to disk or logged. The installer listens only on `127.0.0.1`.

## If the tablet is not found

- Wake/unlock it and reconnect directly rather than through a hub.
- Confirm Developer Mode and SSH access remain enabled.
- Copy the current password again; it can change after reset/recovery.
- For Wi-Fi, replace `10.11.99.1` with the tablet's local IP.
- Do not click past an unexpected host key. Reconnect over USB and verify it.

## After reboot, recovery, and uninstall

Run the installer, click **Find my tablet**, and verify the fingerprint:

- **Reactivate after reboot** starts XOVI/AppLoad without reinstalling files.
- **Restore stock interface** stops the injection and returns to reMarkable.
- **Uninstall TRMNL** removes the app but preserves its data and shared runtime.
- **Uninstall and erase data** also removes TRMNL settings/cache/history/logs.

None of these disables Developer Mode. Use reMarkable's official
[software recovery](https://support.remarkable.com/s/article/Software-recovery)
to leave Developer Mode.

## Windows trust warning

The executable contains icon, manifest, compatibility, and version metadata.
Some releases may still show **Unknown publisher** until the maintainer adds an
Authenticode certificate. A matching SHA-256 proves file integrity against the
release record, but it is not the same as publisher identity. Proceed only if
you trust this project and the checksum matches.
