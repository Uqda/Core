#!/bin/sh
#
# Set FALLBACK_VERSION env var when building from an archive without git tags.
# Example: FALLBACK_VERSION=0.5.13 ./contrib/semver/version.sh

if ! git describe --tags --always 2>/dev/null | grep -q "^v"; then
  fb="${FALLBACK_VERSION:-0.0.0-unknown}"
  case "$*" in
    *--bare*)
      # Match tagged builds: bare output has no leading "v".
      echo "$fb" | sed 's/^v//'
      ;;
    *)
      echo "$fb"
      ;;
  esac
  exit 0
fi

case "$*" in
  *--bare*)
    # Remove the "v" prefix
    git describe --tags --match="v[0-9]*\.[0-9]*\.[0-9]*" | cut -c 2-
    ;;
  *)
    git describe --tags --match="v[0-9]*\.[0-9]*\.[0-9]*"
    ;;
esac
