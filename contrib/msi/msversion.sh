#!/bin/sh

set -eu

TAG=${UQDA_VERSION:-}
if [ -z "$TAG" ]; then
  TAG=$(git describe --abbrev=0 --tags --match="v[0-9]*\.[0-9]*\.[0-9]*" 2>/dev/null || true)
fi

if [ -z "$TAG" ]; then
  # MSI ProductVersion must contain only numeric components.
  COUNT=$(git rev-list --count HEAD 2>/dev/null || printf '0')
  printf '0.0.%d' "$((COUNT % 65535))"
  exit 0
fi

# WiX cannot encode SemVer prerelease labels in ProductVersion. Keep the
# numeric core in MSI metadata; the full beta version remains in the filename.
NUMERIC=$(printf '%s\n' "${TAG#v}" | sed -E 's/^([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
printf '%s\n' "$NUMERIC" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$' || {
  echo "Invalid release version: $TAG" >&2
  exit 1
}

IFS=. read -r MAJOR MINOR PATCH <<EOF
$NUMERIC
EOF
for component in "$MAJOR" "$MINOR" "$PATCH"; do
  [ "$component" -le 65534 ] || {
    echo "MSI version component is too large: $component" >&2
    exit 1
  }
done

printf '%d.%d.%d' "$MAJOR" "$MINOR" "$PATCH"
