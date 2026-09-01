#!/bin/sh

set -eu

TAG=${1:-}
if [ -z "$TAG" ]; then
  TAG=$(
    for NOTES in .github/releases/v*.md; do
      BASENAME=${NOTES##*/}
      printf '%s\n' "${BASENAME%.md}"
    done |
      grep -E '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' |
      awk -F. '
        {
          major = substr($1, 2) + 0
          minor = $2 + 0
          patch = $3 + 0
          if (!found || major > best_major ||
              (major == best_major && minor > best_minor) ||
              (major == best_major && minor == best_minor && patch > best_patch)) {
            found = 1
            best = $0
            best_major = major
            best_minor = minor
            best_patch = patch
          }
        }
        END { if (found) print best }
      '
  )
else
  case "$TAG" in v*) ;; *) TAG="v$TAG" ;; esac
fi

printf '%s\n' "$TAG" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' || {
  echo "Invalid or missing stable release tag: $TAG" >&2
  exit 1
}
test -f ".github/releases/${TAG}.md" || {
  echo "Missing release notes: .github/releases/${TAG}.md" >&2
  exit 1
}

printf '%s\n' "$TAG"
