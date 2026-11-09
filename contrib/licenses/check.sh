#!/bin/sh
# Checks that every third-party Go module compiled into Uqda, on every
# supported operating system, uses an allowed license and is listed in
# THIRD_PARTY_LICENSES.md. Run from the repository root; needs Go.
set -eu

ALLOWED="MIT,BSD-2-Clause,BSD-3-Clause,ISC,Apache-2.0,MPL-2.0,Zlib"
GO_LICENSES="github.com/google/go-licenses/v2@v2.0.1"
LIST=THIRD_PARTY_LICENSES.md

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

go install "$GO_LICENSES"
bin="$(go env GOPATH)/bin/go-licenses"

fail=0
for os in linux windows darwin freebsd openbsd android; do
  echo "== GOOS=$os"
  GOOS=$os "$bin" check ./cmd/... --ignore github.com/Uqda/Core \
    --allowed_licenses="$ALLOWED" || fail=1
  GOOS=$os "$bin" report ./cmd/... --ignore github.com/Uqda/Core \
    >>"$tmp/report.csv" || fail=1
done

# Each reported library must start with a module path listed in the table.
grep -o '^| `[^`]*`' "$LIST" | sed 's/^| `//; s/`$//' >"$tmp/listed"
cut -d, -f1 "$tmp/report.csv" | sort -u | while read -r lib; do
  found=0
  while read -r mod; do
    case "$lib" in "$mod" | "$mod"/*) found=1; break ;; esac
  done <"$tmp/listed"
  if [ "$found" = 0 ]; then
    echo "::error::$lib is compiled into Uqda but missing from $LIST"
    echo missing >>"$tmp/missing"
  fi
done
[ -f "$tmp/missing" ] && fail=1

sort -u "$tmp/report.csv"
exit "$fail"
