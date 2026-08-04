#!/bin/sh
set -u
parent_pid="$1"
device_dir="$2"
original="$3"
disarm="$4"
brightness="$device_dir/brightness"

while kill -0 "$parent_pid" 2>/dev/null; do
    sleep 2
done

if [ ! -e "$disarm" ] && [ -w "$brightness" ]; then
    printf '%s\n' "$original" > "$brightness"
fi
rm -f -- "$disarm"
