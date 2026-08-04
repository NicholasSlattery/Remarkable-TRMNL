# Easy installation guide

No command line is required.

## Before you begin

- Use a reMarkable Paper Pro running firmware 3.26.x or 3.27.x.
- Charge it above 20% and keep it connected during installation.
- In the tablet settings, enable developer/SSH access and display its SSH
  password. The exact menu wording can vary by firmware.
- Connect the tablet directly to the Windows computer with its USB cable.

## Install

1. Extract the downloaded release ZIP.
2. Double-click **TRMNL Installer.exe**. Keep the `payload` folder beside it.
3. Leave the tablet address as `10.11.99.1` for USB. Paste the SSH password.
4. Click **Find my tablet**.
5. Check that the window says **reMarkable Ferrari** and shows supported
   firmware. Confirm the SSH fingerprint.
6. Click **Install TRMNL** and keep the cable connected until it succeeds.
7. On the tablet, open AppLoad and tap TRMNL.

The password is held only in installer memory. It is not saved or logged.

## If the tablet is not found

- Wake and unlock the tablet.
- Reconnect the USB cable directly rather than through a hub.
- Confirm SSH/developer access remains enabled.
- Copy the current password again; the tablet may generate a new one.
- If using Wi-Fi, replace `10.11.99.1` with the tablet's local IP address.

## Recovery and uninstall

Run **TRMNL Installer.exe**, find the tablet, and confirm its fingerprint:

- **Restore stock interface** disables TRMNL and restarts the normal interface
  without deleting settings.
- **Uninstall TRMNL** removes the application but preserves settings, cached
  screens, and the shared extension runtime.

If the installer cannot connect, reboot the tablet. This release deliberately
uses a temporary runtime injection, so a reboot returns to the stock interface.

## Windows warning

Until releases are code-signed, Windows SmartScreen may show an unknown
publisher warning. Verify the ZIP against `SHA256SUMS.txt`, then use **More
info → Run anyway** only when the checksum matches the release page.
