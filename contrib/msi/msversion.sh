#!/bin/sh

# Handle --bare flag (remove "v" prefix from output)
BARE_FLAG=0
case "$*" in
  *--bare*)
    BARE_FLAG=1
    ;;
esac

# Get the version using git describe (more robust than --abbrev=0)
DESCRIBE=$(git describe --tags --match="v[0-9]*\.[0-9]*\.[0-9]*" 2>/dev/null)

# Did getting the version succeed?
if [ $? != 0 ] || [ -z "$DESCRIBE" ]; then
  # Fallback to a valid version when no tags are found
  if [ $BARE_FLAG -eq 1 ]; then
    printf -- "0.0.0.0"
  else
    printf -- "v0.0.0.0"
  fi
  exit 0
fi

# Get the current branch
BRANCH=$(git symbolic-ref -q HEAD --short 2>/dev/null)

# Did getting the branch succeed?
if [ $? != 0 ] || [ -z "$BRANCH" ]; then
  BRANCH="master"
fi

# Extract the tag part (before any dash)
TAG=$(echo "$DESCRIBE" | cut -d "-" -f 1)

# Split out into major, minor and patch numbers
MAJOR=$(echo "$TAG" | cut -c 2- | cut -d "." -f 1)
MINOR=$(echo "$TAG" | cut -c 2- | cut -d "." -f 2)
PATCH=$(echo "$TAG" | cut -c 2- | cut -d "." -f 3 | awk -F"rc" '{print $1}')

# Extract build number if present (commits since tag)
BUILD=0
if echo "$DESCRIBE" | grep -q "-"; then
  BUILD=$(echo "$DESCRIBE" | sed -n 's/.*-\([0-9]*\)-.*/\1/p' | head -1)
  if [ -z "$BUILD" ]; then
    BUILD=$(echo "$DESCRIBE" | sed -n 's/.*-\([0-9]*\)$/\1/p' | head -1)
  fi
  if [ -z "$BUILD" ]; then
    BUILD=0
  fi
fi

# Output in the desired format
if [ $BARE_FLAG -eq 1 ]; then
  # For --bare, output without "v" prefix
  # Ensure we always have a 4-part version for WiX (x.x.x.x)
  if [ $((PATCH)) -eq 0 ]; then
    # If patch is 0, output major.minor.0.build
    printf '%d.%d.0.%d' "$((MAJOR))" "$((MINOR))" "$((BUILD))"
  else
    # If patch is not 0, output major.minor.patch.build
    printf '%d.%d.%d.%d' "$((MAJOR))" "$((MINOR))" "$((PATCH))" "$((BUILD))"
  fi
else
  # Without --bare, output with "v" prefix (original format)
  if [ $((PATCH)) -eq 0 ]; then
    printf 'v%d.%d.0' "$((MAJOR))" "$((MINOR))"
  else
    printf 'v%d.%d.%d' "$((MAJOR))" "$((MINOR))" "$((PATCH))"
  fi
  
  # Add the build tag on non-master branches or if BUILD > 0
  if [ "$BRANCH" != "master" ] || [ $((BUILD)) -gt 0 ]; then
    if [ $((BUILD)) -gt 0 ]; then
      printf -- "-%04d" "$((BUILD))"
    fi
  fi
fi