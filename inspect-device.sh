#!/bin/sh
set -u

value(){ printf '%s=%s\n' "$1" "$2"; }
value timestamp "$(date '+%Y-%m-%dT%H:%M:%S%z')"
value model "$(tr -d '\000' </proc/device-tree/model 2>/dev/null || true)"
value machine "$(cat /sys/devices/soc0/machine 2>/dev/null || true)"
value architecture "$(uname -m)"
value kernel "$(uname -r)"
value os_release "$(grep -E '^(NAME|VERSION|VERSION_ID)=' /etc/os-release 2>/dev/null | tr '\n' ';')"
value remarkable_version "$(grep -E '^(IMG_VERSION|REMARKABLE_RELEASE_VERSION)=' /usr/share/remarkable/update.conf /etc/os-release 2>/dev/null | tr '\n' ';')"
value root_mount "$(findmnt -n -o SOURCE,FSTYPE,OPTIONS / 2>/dev/null || mount | awk '$3=="/"')"
value home_free_kb "$(df -Pk /home/root 2>/dev/null | awk 'NR==2{print $4}')"
value xochitl_active "$(systemctl is-active xochitl 2>/dev/null || true)"
value xochitl_enabled "$(systemctl is-enabled xochitl 2>/dev/null || true)"
value xochitl_pid "$(pidof xochitl 2>/dev/null || true)"
xochitl_pid=$(pidof xochitl 2>/dev/null | awk '{print $1}')
value xochitl_binary "$(readlink -f "/proc/$xochitl_pid/exe" 2>/dev/null || true)"
value display_fb0 "$(cat /sys/class/graphics/fb0/virtual_size 2>/dev/null || true)"
value display_modes "$(cat /sys/class/graphics/fb0/modes 2>/dev/null | tr '\n' ',' || true)"
value rotation "$(cat /sys/class/graphics/fb0/rotate 2>/dev/null || true)"
value xovi_present "$([ -d /home/root/xovi ] && echo yes || echo no)"
value xovi_files "$(find /home/root/xovi -maxdepth 2 -type f \( -name '*.so' -o -name '*.qmd' -o -name '*version*' \) -printf '%p;' 2>/dev/null | head -c 4000)"
value appload_present "$([ -d /home/root/xovi/exthome/appload ] && echo yes || echo no)"
value appload_manifests "$(find /home/root/xovi/exthome/appload -maxdepth 2 \( -name manifest.json -o -name external.manifest.json \) -printf '%p;' 2>/dev/null | head -c 4000)"
value qtfb_shims "$(find /home/root /usr/lib -maxdepth 4 -name 'qtfb*' -printf '%p;' 2>/dev/null | head -c 4000)"
value packages "$(command -v vellum 2>/dev/null || true);$(command -v opkg 2>/dev/null || true);$(command -v toltecctl 2>/dev/null || true)"
value python "$(command -v python3 2>/dev/null || true)"
value curl "$(command -v curl 2>/dev/null || true)"
value wget "$(command -v wget 2>/dev/null || true)"
value dbus "$(command -v busctl 2>/dev/null || true)"
if command -v busctl >/dev/null 2>&1; then value dbus_frontlight "$(busctl --system list 2>/dev/null | grep -Ei 'light|brightness|display' | tr '\n' ';' | head -c 2000)"; fi
for d in /sys/class/backlight/*; do
  [ -d "$d" ] || continue
  value "backlight.$(basename "$d")" "min=$(cat "$d/min_brightness" 2>/dev/null || echo 0),max=$(cat "$d/max_brightness" 2>/dev/null || true),brightness=$(cat "$d/brightness" 2>/dev/null || true),actual=$(cat "$d/actual_brightness" 2>/dev/null || true),type=$(cat "$d/type" 2>/dev/null || true),linear_mapping=$(cat "$d/linear_mapping" 2>/dev/null || true),writable=$([ -w "$d/brightness" ] && echo yes || echo no)"
done
value battery "$(for p in /sys/class/power_supply/*; do [ -d "$p" ] && printf '%s:type=%s,capacity=%s,status=%s;' "$(basename "$p")" "$(cat "$p/type" 2>/dev/null)" "$(cat "$p/capacity" 2>/dev/null)" "$(cat "$p/status" 2>/dev/null)"; done)"
value inputs "$(for p in /sys/class/input/event*/device/name; do [ -r "$p" ] && printf '%s:%s;' "$(basename "$(dirname "$(dirname "$p")")")" "$(cat "$p")"; done | head -c 4000)"
value packages_installed "$(if command -v opkg >/dev/null 2>&1; then opkg list-installed 2>/dev/null | grep -Ei 'xovi|appload|qt-resource|vellum'; elif command -v apk >/dev/null 2>&1; then apk list --installed 2>/dev/null | grep -Ei 'xovi|appload|qt-resource|vellum'; fi | tr '\n' ';' | head -c 4000)"
