#!/bin/sh
# Build the portable Linux archive consumed by linux.sh.
set -eu
[ "$(pwd)" = "$(git rev-parse --show-toplevel)" ] || { echo 'Run from the repository root' >&2; exit 1; }
ARCH=${GOARCH:-$(go env GOARCH)}
case "$ARCH" in amd64|arm64) ;; *) echo "Unsupported archive architecture: $ARCH" >&2; exit 1 ;; esac
VERSION=$(sh contrib/semver/version.sh --bare)
GOOS=linux GOARCH="$ARCH" CGO_ENABLED=0 ./build
STAGE=$(mktemp -d)
trap 'rm -r -- "$STAGE"' EXIT HUP INT TERM
install -m 755 uqda uqdactl "$STAGE/"
install -m 644 contrib/systemd/uqda.service.quick "$STAGE/uqda.service"
install -m 755 contrib/packaging/install-config.sh "$STAGE/install-config.sh"
tar -C "$STAGE" -czf "uqda-${VERSION}-linux-${ARCH}.tar.gz" uqda uqdactl uqda.service install-config.sh
