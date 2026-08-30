#!/bin/sh

# Fetches the installer belonging to the requested UQDA release, verifies it
# against that release's SHA256SUMS, and runs the normal upgrade path.

set -eu

REPOSITORY=${UQDA_REPOSITORY:-Uqda/Core}
VERSION=${UQDA_VERSION:-}
case "${1:-}" in
	--version) [ "$#" -ge 2 ] || { echo "--version needs a value" >&2; exit 1; }; VERSION=$2; shift 2 ;;
esac
TMPDIR_UQDA=$(mktemp -d "${TMPDIR:-/tmp}/uqda-update.XXXXXX")
trap 'rm -rf "$TMPDIR_UQDA"' EXIT HUP INT TERM

fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget --https-only --tries=3 --output-document="$2" "$1"
	else
		echo "curl or wget is required" >&2; exit 1
	fi
}

if [ -z "$VERSION" ]; then
	fetch "https://raw.githubusercontent.com/$REPOSITORY/main/VERSION" "$TMPDIR_UQDA/VERSION"
	VERSION=$(sed -n '1p' "$TMPDIR_UQDA/VERSION" | tr -d '\r\n')
fi
case "$VERSION" in
	v[0-9]*.[0-9]*.[0-9]*) ;;
	*) echo "invalid release tag: $VERSION" >&2; exit 1 ;;
esac
BASE_URL=${UQDA_RELEASE_BASE_URL:-https://github.com/$REPOSITORY/releases/download/$VERSION}

fetch "$BASE_URL/SHA256SUMS" "$TMPDIR_UQDA/SHA256SUMS"
fetch "$BASE_URL/install.sh" "$TMPDIR_UQDA/install.sh"
expected=$(awk '$2 == "install.sh" || $2 == "./install.sh" { print $1; exit }' "$TMPDIR_UQDA/SHA256SUMS")
[ -n "$expected" ] || { echo "install.sh is missing from SHA256SUMS" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "$TMPDIR_UQDA/install.sh" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
	actual=$(shasum -a 256 "$TMPDIR_UQDA/install.sh" | awk '{print $1}')
else
	actual=$(openssl dgst -sha256 "$TMPDIR_UQDA/install.sh" | awk '{print $NF}')
fi
[ "$actual" = "$expected" ] || { echo "install.sh checksum mismatch" >&2; exit 1; }

UQDA_VERSION=$VERSION UQDA_RELEASE_BASE_URL=$BASE_URL sh "$TMPDIR_UQDA/install.sh" --update "$@"
