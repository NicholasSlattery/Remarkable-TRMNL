# Privacy and data handling

TRMNL for reMarkable has no maintainer-operated telemetry or analytics. It
connects directly from the tablet to the configured TRMNL cloud or BYOS origin.
That service's own privacy terms apply to dashboard/plugin content and requests.

## Data stored on the tablet

- Configuration and Device API key:
  `/home/root/.config/trmnl-remarkable/config.json`
- Downloaded screen cache: `/home/root/.cache/trmnl-remarkable/`
- Refresh history, logs, diagnostics state, and battery test:
  `/home/root/.local/share/trmnl-remarkable/`
- Application: `/home/root/xovi/exthome/appload/trmnl-remarkable/`

The configuration file is owner-only (`0600`). API keys are masked after save,
not returned to QML, and redacted from diagnostics. Tablet SSH passwords remain
in the localhost installer's browser memory for the current process and are not
saved or logged by the installer.

## Network data

Device API requests can send the device-scoped access token, configured device
ID/MAC, app version, model name, battery voltage, and (if later implemented by
the platform) signal strength. The server returns image URLs and refresh
instructions. Firmware/reset instructions are ignored.

Production API and image traffic requires HTTPS. The only HTTP exception is a
loopback development mock on the tablet. Credential-bearing redirects cannot
change origin or protocol.

## Update check

**Check GitHub for a newer TRMNL version** is off by default. It is the only
request the app makes to a host other than your configured dashboard server.
When you turn it on, the tablet asks `api.github.com` once a day for the newest
published release tag. That request carries no API key, device ID, or dashboard
content, but it does reveal the tablet's IP address to GitHub, and GitHub's own
terms apply to it. Turning the setting off stops the request immediately.

The Wi-Fi signal strength the app reports to TRMNL alongside battery voltage is
read locally from the wireless interface. It is sent only to your configured
dashboard server, in the same request as the rest of the device metadata.

## Removal

**Uninstall TRMNL** preserves configuration/cache for reinstall. **Uninstall and
erase data** removes the listed TRMNL data directories. Shared XOVI/AppLoad files
are preserved because other extensions may rely on them. Neither action removes
data already sent to TRMNL/BYOS or disables reMarkable Developer Mode.
