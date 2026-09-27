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
    ANNUAL=${BASE%.0}
    ANNUAL=${ANNUAL%.0}
    TITLE="Uqda $ANNUAL"
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
    # The public GA tag has an explicit .0 patch, but MSI permits only three
    # numeric fields. Keep the established 26.0.1000 installer sequence.
    INSTALLBASE=$BASE
    case "$BASE" in *.*.*) INSTALLBASE=${BASE%.*} ;; esac
    printf '%s.%s\n' "$INSTALLBASE" "$REV" ;;
  '') printf 'v%s\n' "$VERSION" ;;
  *) echo "Unknown version format: $1" >&2; exit 1 ;;
esac
