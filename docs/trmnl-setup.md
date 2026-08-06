# TRMNL BYOD setup for reMarkable Paper Pro

Use this guide when connecting the app to the hosted TRMNL cloud. If you run a
fully self-hosted BYOS server, the hosted BYOD license and TRMNL Device API key
are not required; enter your HTTPS BYOS origin and its device identity instead.

## 1. Obtain and claim a BYOD license

The hosted TRMNL service requires one BYOD license for a third-party device.

[Buy TRMNL BYOD](https://shop.trmnl.com/products/byod)

This is a direct, non-affiliate product link; the project receives no commission.

After purchase, sign in to TRMNL and use the order number to claim/add the BYOD
device. TRMNL's current help center explains the claim and Friendly ID flow:
[Find your Friendly ID](https://help.trmnl.com/en/articles/12632379-find-your-friendly-id).

## 2. Configure the device model

In the TRMNL device settings, choose or create the closest custom model with:

- Resolution: **1620 x 2160**
- Orientation: **portrait**
- Color capability: **full color / 16.7M** when the model editor offers it
- Image format: **PNG**

Use fit mode in the tablet app if a plugin sends a different aspect ratio. The
Paper Pro can display color, but a plugin/template must also render color.

## 3. Get the correct Device API key

1. Sign in at [trmnl.com](https://trmnl.com/).
2. Open **Devices** and select/edit the claimed BYOD device.
3. Open **Developer Perks** (the label may be under a gear/edit view).
4. Copy that device's **Device API Key**.

Do not use the Account API token from account settings. Account tokens begin
with `user_`, use bearer authentication, and are for a different API. This app
needs the device-scoped key used as the `access-token` header. Treat it as a
password and never paste it into an issue, screenshot, log, or diagnostics post.

## 4. Enter it on the tablet

1. Open TRMNL in AppLoad and tap the upper-right hotspot.
2. Choose **Settings** and leave the server on **TRMNL cloud**.
3. Paste the Device API key.
4. Confirm the pre-filled **Device ID / MAC address** matches the Wi-Fi MAC in
   the TRMNL BYOD device record. Correct it if the account uses another value.
5. Tap **Test connection**, then **Save**.

The key is stored at `/home/root/.config/trmnl-remarkable/config.json` with mode
`0600`. It is masked in the UI after save and redacted from diagnostics.

## 5. Confirm operation

Tap **Next screen** and confirm a playlist image appears. In diagnostics, check
that the last refresh succeeded and that the next refresh time is present.

An HTTP 401/403 usually means the wrong key or unclaimed device. A `user_...`
value is definitely the wrong key. An HTTP 429 is rate limiting; the client honors the
server's `Retry-After` value instead of repeatedly requesting.
