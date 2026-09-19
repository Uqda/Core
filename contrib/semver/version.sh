#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
VERSION=$(cat "$ROOT/src/version/VERSION")
case "$VERSION" in
  *[!0-9a-z.-]*|'') echo "Invalid release version" >&2; exit 1 ;;
esac
BASE=${VERSION%%-*}
case "${1:-}" in
  --bare) printf '%s\n' "$VERSION" ;;
  --debian) printf '%s\n' "$VERSION" | sed 's/-/~/g' ;;
  --title|--display)
    TITLE="Uqda ${BASE%.0}"
    case "$VERSION" in
      *-beta.*) TITLE="$TITLE Beta ${VERSION##*-beta.}" ;;
      *-*) TITLE="$TITLE ${VERSION#*-}" ;;
    esac
    if [ "$1" = --display ]; then TITLE="Uqda Core ${TITLE#Uqda }"; fi
    printf '%s\n' "$TITLE" ;;
  --prerelease) case "$VERSION" in *-*) echo true ;; *) echo false ;; esac ;;
  --installer)
    # Numeric installer ordering: beta 1..999, general release 1000.
    case "$VERSION" in *-beta.*) REV=${VERSION##*-beta.} ;; *-*) exit 1 ;; *) REV=1000 ;; esac
    printf '%s.%s\n' "$BASE" "$REV" ;;
  '') printf 'v%s\n' "$VERSION" ;;
  *) echo "Unknown version format: $1" >&2; exit 1 ;;
esac
