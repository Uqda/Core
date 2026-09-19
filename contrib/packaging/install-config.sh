#!/bin/sh
# Initialize or copy persistent identity without overwriting an existing config.
set -eu
umask 077
BIN=${UQDA_BIN:-uqda}
DEST=${1:?configuration destination required}
shift
if [ -e "$DEST" ]; then
  "$BIN" -useconffile "$DEST" -checkconf
  exit 0
fi
SOURCE=
for candidate in "$@"; do
  if [ -e "$candidate" ]; then
    if [ -n "$SOURCE" ]; then
      echo "Multiple legacy configurations found; select and migrate one manually." >&2
      exit 1
    fi
    SOURCE=$candidate
  fi
done
mkdir -p "$(dirname "$DEST")"
TEMP=$(mktemp "${DEST}.tmp.XXXXXX")
trap 'rm -f "$TEMP"' EXIT HUP INT TERM
if [ -n "$SOURCE" ]; then
  "$BIN" -useconffile "$SOURCE" -checkconf
  cp "$SOURCE" "$TEMP"
else
  "$BIN" -genconf > "$TEMP"
fi
"$BIN" -useconffile "$TEMP" -checkconf
# A hard link publishes complete bytes without clobbering a concurrent install.
ln "$TEMP" "$DEST"
