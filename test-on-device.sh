#!/bin/sh
set -eu
ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
fail=0
check(){ if "$@"; then echo "PASS: $*"; else echo "FAIL: $*"; fail=1; fi; }
check test -x "$ROOT/dist/trmnl-remarkable-app/backend/entry"
check test -s "$ROOT/dist/trmnl-remarkable-app/resources.rcc"
check test -f /home/root/xovi/exthome/appload/trmnl-remarkable/manifest.json
check systemctl is-active --quiet xochitl
check test "$(stat -c %a /home/root/.config/trmnl-remarkable/config.json 2>/dev/null || echo 600)" = 600
check sh -c 'for d in /sys/class/backlight/*; do [ -r "$d/brightness" ] && [ -r "$d/max_brightness" ] && exit 0; done; exit 1'
exit "$fail"
