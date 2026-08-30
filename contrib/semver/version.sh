#!/bin/sh

set -e

TAG=$(git describe --tags --match="v[0-9]*\.[0-9]*\.[0-9]*" 2>/dev/null || true)
if [ -z "$TAG" ]; then
  # Untagged development builds still need a deterministic, valid version for
  # archives and native package metadata.
  COUNT=$(git rev-list --count HEAD 2>/dev/null || printf '0')
  TAG="v0.0.0-dev.${COUNT}"
fi

case "$*" in
  *--bare*) printf '%s' "${TAG#v}" ;;
  *) printf '%s' "$TAG" ;;
esac
