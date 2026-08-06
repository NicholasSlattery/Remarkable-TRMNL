#!/bin/sh
set -u
parent_pid="$1"
device_dir="$2"
original="$3"
disarm="$4"
brightness="$device_dir/brightness"

# Ten seconds is frequent enough that the front light is restored before the
# user notices, and infrequent enough that this watchdog is not itself a
# measurable drain on a tablet meant to sit idle for days.
while kill -0 "$parent_pid" 2>/dev/null; do
    sleep 10
done

if [ ! -e "$disarm" ] && [ -w "$brightness" ]; then
    printf '%s\n' "$original" > "$brightness"
fi
rm -f -- "$disarm"
