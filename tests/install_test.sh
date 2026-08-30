#!/bin/sh

set -eu
ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)

plan() {
	UQDA_TEST_OS=$1 UQDA_TEST_ARCH=$2 UQDA_TEST_PLATFORM=$3 sh "$ROOT/install.sh" --dry-run
}

assert_asset() {
	output=$(plan "$1" "$2" "$3")
	printf '%s\n' "$output" | grep -F "asset=$4 " >/dev/null || {
		printf 'wrong plan for %s/%s/%s:\n%s\n' "$1" "$2" "$3" "$output" >&2
		exit 1
	}
}

assert_asset Linux x86_64 systemd uqda-0.1.0-beta.2-amd64.deb
assert_asset Linux aarch64 systemd uqda-0.1.0-beta.2-arm64.deb
assert_asset Linux armv7l portable uqda-v0.1.0-beta.2-linux-armv7.tar.gz
assert_asset Linux mipsel edgeos2x uqda-edgeos2x-0.1.0-beta.2-mipsel.deb
assert_asset Linux x86_64 vyos13 uqda-vyos13-0.1.0-beta.2-amd64.deb
assert_asset Darwin arm64 launchd uqda-0.1.0-beta.2-macos-arm64.pkg
assert_asset FreeBSD amd64 portable uqda-v0.1.0-beta.2-freebsd-amd64.tar.gz
assert_asset OpenBSD arm64 portable uqda-v0.1.0-beta.2-openbsd-arm64.tar.gz

if UQDA_TEST_OS=Linux UQDA_TEST_ARCH=riscv64 UQDA_TEST_PLATFORM=portable sh "$ROOT/install.sh" --dry-run >/dev/null 2>&1; then
	echo "unsupported architecture was accepted" >&2
	exit 1
fi

echo "installer platform matrix passed"
