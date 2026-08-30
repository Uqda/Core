#!/bin/sh

set -eu

TAG=${UQDA_VERSION:-}
if [ -n "$TAG" ]; then
  printf '%s\n' "$TAG" | grep -Eq '^v?[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$' || {
    echo "Invalid UQDA_VERSION: $TAG" >&2
    exit 1
  }
  case "$TAG" in v*) ;; *) TAG="v$TAG" ;; esac
else
  TAG=$(git describe --tags --match="v[0-9]*\.[0-9]*\.[0-9]*" 2>/dev/null || true)
fi
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
