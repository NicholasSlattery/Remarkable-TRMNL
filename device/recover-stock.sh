#!/bin/sh
set -eu
APP_DEST="/home/root/xovi/exthome/appload/trmnl-remarkable"
STATE="/home/root/.local/share/trmnl-remarkable"
DISABLED="$STATE/recovery/trmnl-remarkable.$(date +%Y%m%d-%H%M%S).disabled"
DROPIN="/etc/systemd/system/xochitl.service.d"
disabled_path=""
trmnl_pids() {
  for proc in /proc/[0-9]*; do
    [ -e "$proc/exe" ] || continue
    exe=$(readlink -f "$proc/exe" 2>/dev/null || true)
    [ "$exe" = "$APP_DEST/backend/entry" ] && printf '%s\n' "${proc##*/}"
  done
}
for pid in $(trmnl_pids); do kill -TERM "$pid" 2>/dev/null || true; done
sleep 2
for pid in $(trmnl_pids); do kill -KILL "$pid" 2>/dev/null || true; done
if [ -d "$APP_DEST" ]; then
  mkdir -p "$STATE/recovery"
  mv "$APP_DEST" "$DISABLED"
  disabled_path="$DISABLED"
fi
# XOVI's service drop-in is a temporary exact-path mount. Removing it and
# restarting Xochitl restores the vendor unit without changing the read-only
# system partition. A reboot has the same effect.
if mount | grep -F " on $DROPIN type " >/dev/null 2>&1; then
  umount "$DROPIN"
fi
systemctl daemon-reload
systemctl restart xochitl
systemctl is-active --quiet xochitl
# Xochitl needs a moment to appear in /proc after the restart. Without this
# wait an empty PID makes the injection probe silently report success.
xochitl_pid=""
attempt=0
while [ "$attempt" -lt 20 ]; do
  xochitl_pid=$(pidof xochitl 2>/dev/null | awk '{print $1}')
  if [ -n "$xochitl_pid" ] && [ -r "/proc/$xochitl_pid/environ" ]; then
    break
  fi
  attempt=$((attempt + 1))
  sleep 1
done
if [ -z "$xochitl_pid" ] || [ ! -r "/proc/$xochitl_pid/environ" ]; then
  echo "Recovery verification failed: Xochitl did not report a readable process after restart" >&2
  exit 41
fi
if tr '\000' '\n' <"/proc/$xochitl_pid/environ" | grep -q '^LD_PRELOAD=/home/root/xovi/xovi.so$'; then
  echo "Recovery verification failed: XOVI is still injected" >&2
  exit 40
fi
if [ -n "$disabled_path" ]; then
  echo "Stock reMarkable interface is active without XOVI injection; TRMNL is disabled at $disabled_path"
else
  echo "Stock reMarkable interface is active without XOVI injection; no installed TRMNL bundle was found"
fi
