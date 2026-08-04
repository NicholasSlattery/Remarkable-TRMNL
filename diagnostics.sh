#!/bin/sh
set -u
echo "TRMNL reMarkable diagnostics"
echo "timestamp=$(date '+%Y-%m-%dT%H:%M:%S%z')"
echo "model=$(tr -d '\000' </proc/device-tree/model 2>/dev/null || true)"
echo "machine=$(cat /sys/devices/soc0/machine 2>/dev/null || true)"
echo "arch=$(uname -m)"
echo "kernel=$(uname -r)"
echo "os_version=$(grep -E '^(VERSION|VERSION_ID)=' /etc/os-release 2>/dev/null | tr '\n' ' ')"
echo "display=$(cat /sys/class/graphics/fb0/virtual_size 2>/dev/null || echo unavailable)"
echo "xochitl=$(systemctl is-active xochitl 2>/dev/null || true)"
echo "xovi=$([ -d /home/root/xovi ] && echo present || echo absent)"
echo "appload=$([ -d /home/root/xovi/exthome/appload ] && echo present || echo absent)"
echo "trmnl_app=$([ -d /home/root/xovi/exthome/appload/trmnl-remarkable ] && echo installed || echo absent)"
for d in /sys/class/backlight/*; do
  [ -d "$d" ] || continue
  echo "backlight=$(basename "$d") min=$(cat "$d/min_brightness" 2>/dev/null || echo 0) max=$(cat "$d/max_brightness" 2>/dev/null || echo unknown) current=$(cat "$d/brightness" 2>/dev/null || echo unknown)"
done
battery=unknown
for supply in /sys/class/power_supply/*; do
  [ "$(cat "$supply/type" 2>/dev/null || true)" = Battery ] || continue
  [ -f "$supply/capacity" ] || continue
  battery=$(cat "$supply/capacity")
  break
done
echo "battery=$battery"
backend_pids=""
for proc in /proc/[0-9]*; do
  [ "$(readlink -f "$proc/exe" 2>/dev/null || true)" = "/home/root/xovi/exthome/appload/trmnl-remarkable/backend/entry" ] && backend_pids="$backend_pids${proc##*/},"
done
echo "backend_pids=$backend_pids"
echo "disk_free_kb=$(df -Pk /home/root 2>/dev/null | awk 'NR==2 {print $4}')"
