#!/bin/sh
set -eu

STAGE=/tmp/trmnl-install
PROJECT=/home/root/trmnl-remarkable
STATE=/home/root/.local/share/trmnl-remarkable
created_xovi=0
installed_appload=0
completed=0
rollback_dir="$STAGE/rollback"
project_backup=""
project_created=0

rollback() {
    [ "$completed" -eq 1 ] && return 0
    if [ "$installed_appload" -eq 1 ]; then
        for item in appload.so qtfb-shim.so qtfb-shim-32bit.so; do
            case "$item" in
                appload.so) destination=/home/root/xovi/extensions.d/appload.so ;;
                *) destination="/home/root/shims/$item" ;;
            esac
            if [ -f "$rollback_dir/$item" ]; then
                cp "$rollback_dir/$item" "$destination"
            else
                rm -f -- "$destination"
            fi
        done
    fi
    if [ "$created_xovi" -eq 1 ]; then
        rm -rf -- /home/root/xovi
    fi
    if [ "$project_created" -eq 1 ]; then
        rm -rf -- "$PROJECT"
        [ -z "$project_backup" ] || mv "$project_backup" "$PROJECT"
    fi
    echo "Installation failed; newly installed runtime files were rolled back." >&2
}
trap rollback EXIT HUP INT TERM

expect_hash() {
    expected="$1"; file="$2"
    actual=$(sha256sum "$file" | awk '{print $1}')
    [ "$actual" = "$expected" ] || { echo "Checksum mismatch: $file" >&2; exit 30; }
}

model=$(tr -d '\000' </proc/device-tree/model 2>/dev/null || true)
arch=$(uname -m)
os_version=$(sed -n 's/^IMG_VERSION="\([^"]*\)"/\1/p' /etc/os-release)
[ "$model" = "reMarkable Ferrari" ] || { echo "Not a Paper Pro: $model" >&2; exit 31; }
[ "$arch" = "aarch64" ] || { echo "Unsupported architecture: $arch" >&2; exit 32; }
case "$os_version" in 3.26.*|3.27.*) ;; *) echo "AppLoad 0.5.3 is incompatible with OS $os_version (requires >=3.26,<3.28)" >&2; exit 33;; esac
[ "$(systemctl is-active xochitl)" = active ] || { echo "Xochitl is not active" >&2; exit 34; }

expect_hash 32d64d1262ddc984e3235c7d0340a398fe6d5b3efa6a979865f5977b32630d27 "$STAGE/xovi-aarch64.tar.gz"
source_hash=$(tr -d '[:space:]' <"$STAGE/trmnl-remarkable-device.sha256")
case "$source_hash" in *[!0-9a-f]*|'') echo "Invalid source checksum manifest" >&2; exit 30;; esac
[ "${#source_hash}" -eq 64 ] || { echo "Invalid source checksum length" >&2; exit 30; }
expect_hash "$source_hash" "$STAGE/trmnl-remarkable-device.tar.gz"
expect_hash 31214cbbe64c8bfe7d99096f077c3009dba8a42ef1a733801aa0ec59c134e7cc "$STAGE/appload/appload.so"
expect_hash 6df704049aa057ff6374eaaa03a4f4a4d683b7c1ce772920d1a124be74d782c4 "$STAGE/appload/qtfb-shim.so"
expect_hash aa4fb1e6f2edf5ef0137360cac77713a24ab508800301f81c19c579fee3f5031 "$STAGE/appload/qtfb-shim-32bit.so"

mkdir -p "$STATE/install-backup" "$STATE/upstream-licenses"
{
    echo "timestamp=$(date '+%Y-%m-%dT%H:%M:%S%z')"
    echo "model=$model"
    echo "os_version=$os_version"
    echo "architecture=$arch"
    echo "xochitl=$(systemctl is-active xochitl)"
    echo "root_mount=$(mount | awk '$3=="/"{print $0}')"
    echo "home_free_kb=$(df -Pk /home/root | awk 'NR==2{print $4}')"
    echo "brightness=$(cat /sys/class/backlight/rm_frontlight/brightness)"
} > "$STATE/install-backup/preinstall-state.txt"
chmod 600 "$STATE/install-backup/preinstall-state.txt"

if [ ! -e /home/root/xovi ]; then
    tar -tzf "$STAGE/xovi-aarch64.tar.gz" | awk '/^\// || /(^|\/)\.\.($|\/)/ || $0 !~ /^xovi\// { bad=1 } END { exit bad }' || { echo "Unsafe XOVI archive" >&2; exit 35; }
    created_xovi=1
    tar -xzf "$STAGE/xovi-aarch64.tar.gz" -C /home/root
fi
[ -f /home/root/xovi/xovi.so ] && [ -f /home/root/xovi/extensions.d/qt-resource-rebuilder.so ] || { echo "Compatible XOVI runtime is missing or incomplete" >&2; exit 36; }
expect_hash d4df820c25c634c511de11067279d8310fa4f656dc52bd4540db6beac4ffd446 /home/root/xovi/xovi.so
expect_hash 6726f561557406f36347e43fc2b44a88deef4fb273d2ece88f48f427dad8800f /home/root/xovi/extensions.d/qt-resource-rebuilder.so

mkdir -p /home/root/xovi/exthome/appload /home/root/shims
mkdir -p "$rollback_dir"
[ ! -f /home/root/xovi/extensions.d/appload.so ] || cp /home/root/xovi/extensions.d/appload.so "$rollback_dir/appload.so"
[ ! -f /home/root/shims/qtfb-shim.so ] || cp /home/root/shims/qtfb-shim.so "$rollback_dir/qtfb-shim.so"
[ ! -f /home/root/shims/qtfb-shim-32bit.so ] || cp /home/root/shims/qtfb-shim-32bit.so "$rollback_dir/qtfb-shim-32bit.so"
installed_appload=1
cp "$STAGE/appload/appload.so" /home/root/xovi/extensions.d/appload.so
cp "$STAGE/appload/qtfb-shim.so" /home/root/shims/qtfb-shim.so
cp "$STAGE/appload/qtfb-shim-32bit.so" /home/root/shims/qtfb-shim-32bit.so
chmod 644 /home/root/xovi/extensions.d/appload.so /home/root/shims/qtfb-shim.so /home/root/shims/qtfb-shim-32bit.so
cp "$STAGE/licenses/XOVI-LICENSE" "$STATE/upstream-licenses/XOVI-LICENSE"
cp "$STAGE/licenses/APPLOAD-LICENSE" "$STATE/upstream-licenses/APPLOAD-LICENSE"
cp "$STAGE/licenses/EXTENSIONS-LICENSE" "$STATE/upstream-licenses/EXTENSIONS-LICENSE"
printf '%s\n' 'xovi_appload_runtime=installed-by-trmnl-remarkable' > "$STATE/install-backup/runtime-owned"
chmod 600 "$STATE/install-backup/runtime-owned"

tar -tzf "$STAGE/trmnl-remarkable-device.tar.gz" | awk '/^\// || /(^|\/)\.\.($|\/)/ { bad=1 } END { exit bad }' || { echo "Unsafe TRMNL archive" >&2; exit 37; }
if [ -e "$PROJECT" ]; then
    project_backup="$STATE/install-backup/source.$(date '+%Y%m%d-%H%M%S')"
    mv "$PROJECT" "$project_backup"
fi
mkdir -p "$PROJECT"
project_created=1
tar -xzf "$STAGE/trmnl-remarkable-device.tar.gz" -C "$PROJECT"
chmod 755 "$PROJECT"/*.sh "$PROJECT"/dist/trmnl-remarkable-app/scripts/brightness_guard.sh "$PROJECT"/dist/trmnl-remarkable-app/backend/entry
chmod 755 "$PROJECT"/scripts/*.sh 2>/dev/null || true
sh "$PROJECT/install.sh"

completed=1
trap - EXIT HUP INT TERM

echo "RUNTIME_INSTALL_OK"
echo "model=$model os=$os_version arch=$arch"
echo "xovi=0.3.3 extensions=19.0.0 appload=0.5.3"
