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

## Removal

**Uninstall TRMNL** preserves configuration/cache for reinstall. **Uninstall and
erase data** removes the listed TRMNL data directories. Shared XOVI/AppLoad files
are preserved because other extensions may rely on them. Neither action removes
data already sent to TRMNL/BYOS or disables reMarkable Developer Mode.
