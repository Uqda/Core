#!/bin/sh

set -eu
ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TEST_VERSION=${UQDA_TEST_VERSION:-v0.1.1}
BARE_VERSION=${TEST_VERSION#v}

plan() {
	UQDA_VERSION=$TEST_VERSION UQDA_TEST_OS=$1 UQDA_TEST_ARCH=$2 UQDA_TEST_PLATFORM=$3 sh "$ROOT/install.sh" --dry-run
}

assert_asset() {
	output=$(plan "$1" "$2" "$3")
	printf '%s\n' "$output" | grep -F "asset=$4 " >/dev/null || {
		printf 'wrong plan for %s/%s/%s:\n%s\n' "$1" "$2" "$3" "$output" >&2
		exit 1
	}
}

assert_asset Linux x86_64 systemd-deb "uqda-${BARE_VERSION}-amd64.deb"
assert_asset Linux aarch64 systemd-deb "uqda-${BARE_VERSION}-arm64.deb"
assert_asset Linux x86_64 systemd-portable "uqda-${TEST_VERSION}-linux-amd64.tar.gz"
assert_asset Linux armv7l systemd-portable "uqda-${TEST_VERSION}-linux-armv7.tar.gz"
assert_asset Linux mipsel edgeos2x "uqda-edgeos2x-${BARE_VERSION}-mipsel.deb"
assert_asset Linux x86_64 vyos13 "uqda-vyos13-${BARE_VERSION}-amd64.deb"
assert_asset Darwin arm64 launchd "uqda-${BARE_VERSION}-macos-arm64-unsigned.pkg"
assert_asset FreeBSD amd64 portable "uqda-${TEST_VERSION}-freebsd-amd64.tar.gz"
assert_asset OpenBSD arm64 portable "uqda-${TEST_VERSION}-openbsd-arm64.tar.gz"

if UQDA_VERSION=$TEST_VERSION UQDA_TEST_OS=Linux UQDA_TEST_ARCH=riscv64 UQDA_TEST_PLATFORM=portable sh "$ROOT/install.sh" --dry-run >/dev/null 2>&1; then
	echo "unsupported architecture was accepted" >&2
	exit 1
fi

fixture=$(mktemp "$ROOT/uqda-releases.XXXXXX")
trap 'rm -f "$fixture"' EXIT HUP INT TERM
printf '%s\n' '[' '  {' '    "tag_name": "v9.8.7",' '    "prerelease": false' '  }' ']' > "$fixture"
resolved=$(UQDA_VERSION= UQDA_RELEASES_FILE=$fixture UQDA_RESOLVE_ONLY=1 sh "$ROOT/updater.sh")
[ "$resolved" = v9.8.7 ] || { echo "automatic release discovery failed" >&2; exit 1; }

grep -F 'chmod 0600 "$CONFIG_FILE"' "$ROOT/install.sh" >/dev/null
grep -F 'systemd-deb) SOCKET_PATH=/var/run/uqda/uqda.sock' "$ROOT/install.sh" >/dev/null
grep -F 'systemd-portable) SOCKET_PATH=/var/run/uqda.sock' "$ROOT/install.sh" >/dev/null

echo "installer platform matrix passed for $TEST_VERSION"
