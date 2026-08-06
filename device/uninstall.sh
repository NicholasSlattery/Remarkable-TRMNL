#!/bin/sh
set -eu
APP_DEST="/home/root/xovi/exthome/appload/trmnl-remarkable"
PROJECT="/home/root/trmnl-remarkable"
STATE="/home/root/.local/share/trmnl-remarkable"
DROPIN="/etc/systemd/system/xochitl.service.d"
stop_trmnl_backend() {
  for proc in /proc/[0-9]*; do
    [ -e "$proc/exe" ] || continue
    exe=$(readlink -f "$proc/exe" 2>/dev/null || true)
    [ "$exe" = "$APP_DEST/backend/entry" ] || continue
    kill -TERM "${proc##*/}" 2>/dev/null || true
  done
}
stop_trmnl_backend
sleep 2
rm -rf -- "$APP_DEST"
if mount | grep -F " on $DROPIN type " >/dev/null 2>&1; then
  umount "$DROPIN"
fi
systemctl daemon-reload

rm -rf -- "$STATE/recovery" "$STATE/upstream-licenses"
rm -f -- "$STATE/install-backup/runtime-owned"
if [ "${1:-}" = "--purge" ]; then
  auth=/home/root/.ssh/authorized_keys
  if [ -f "$auth" ]; then
    tmp="$auth.trmnl-uninstall.$$"
    grep -v ' trmnl-remarkable-codex$' "$auth" >"$tmp" || true
    chmod 600 "$tmp"
    mv "$tmp" "$auth"
  fi
  rm -rf -- /home/root/.config/trmnl-remarkable /home/root/.cache/trmnl-remarkable "$STATE"
fi
rm -rf -- "$PROJECT"
# Restart after this SSH command has returned. On this firmware Xochitl's
# startup password maintenance can close an active Dropbear session.
nohup sh -c 'sleep 2; systemctl restart xochitl' >/tmp/trmnl-uninstall-restart.log 2>&1 </dev/null &
echo "TRMNL was removed. The shared XOVI/AppLoad runtime was preserved so other extensions cannot be deleted accidentally. Settings/cache were preserved unless --purge was specified; stock Xochitl is restarting."
