#!/bin/sh
set -eu

PROJECT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
APP_SOURCE="$PROJECT_DIR/dist/trmnl-remarkable-app"
APP_ROOT="/home/root/xovi/exthome/appload"
APP_DEST="$APP_ROOT/trmnl-remarkable"
BACKUP_DIR="/home/root/.local/share/trmnl-remarkable/install-backup"
STAGED="$APP_ROOT/.trmnl-remarkable.new.$$"

model=$(tr -d '\000' </proc/device-tree/model 2>/dev/null || true)
machine=$(cat /sys/devices/soc0/machine 2>/dev/null || true)
case "$model|$machine" in
  "reMarkable Ferrari|reMarkable Ferrari"|*"Paper Pro"*) ;;
  *) echo "Refusing installation: connected device is not identified as a reMarkable Paper Pro ($model $machine)" >&2; exit 20 ;;
esac
os_version=$(sed -n 's/^IMG_VERSION="\{0,1\}\([^" ]*\)"\{0,1\}$/\1/p' /etc/os-release | head -n1)
case "$os_version" in 3.26.*|3.27.*) ;; *) echo "Refusing installation: unsupported reMarkable OS $os_version (supported: 3.26.x and 3.27.x)" >&2; exit 25;; esac

[ -d /home/root/xovi ] || { echo "Compatible XOVI is not installed; inspect OS compatibility before installing it." >&2; exit 21; }
[ -d "$APP_ROOT" ] || { echo "AppLoad application directory is missing: $APP_ROOT" >&2; exit 22; }
[ -f "$APP_SOURCE/manifest.json" ] && [ -f "$APP_SOURCE/resources.rcc" ] && [ -x "$APP_SOURCE/backend/entry" ] || { echo "Built AppLoad bundle is incomplete" >&2; exit 23; }

free_kb=$(df -Pk /home/root | awk 'NR==2 {print $4}')
[ "${free_kb:-0}" -ge 51200 ] || { echo "At least 50 MiB free under /home/root is required" >&2; exit 24; }

mkdir -p "$BACKUP_DIR" "$APP_ROOT"
rm -rf -- "$STAGED"
trap 'rm -rf -- "$STAGED"' EXIT HUP INT TERM
cp -R "$APP_SOURCE" "$STAGED"
chmod 0755 "$STAGED/backend/entry" "$STAGED/scripts/brightness_guard.sh"
chmod 0644 "$STAGED/manifest.json" "$STAGED/resources.rcc" "$STAGED/icon.png"
chown -R root:root "$STAGED"

"$STAGED/backend/entry" --self-check "$STAGED"

if [ -e "$APP_DEST" ]; then
  stamp=$(date +%Y%m%d-%H%M%S)
  mv "$APP_DEST" "$BACKUP_DIR/trmnl-remarkable.$stamp"
fi
mv "$STAGED" "$APP_DEST"
trap - EXIT HUP INT TERM

echo "TRMNL AppLoad bundle installed: $APP_DEST"
echo "AppLoad is starting. When the interface returns, open AppLoad and tap TRMNL."
nohup sh -c 'sleep 2; /home/root/xovi/start' >/tmp/trmnl-xovi-start.log 2>&1 </dev/null &
