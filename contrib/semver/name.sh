#!/bin/sh
#
# Produces a build name based on the current git branch.
# Requires: git describe --tags with a vX.Y.Z tag reachable.
# Without a git tag, output may be unexpected (e.g. "uqda-untagged").
# When building from a source archive without .git, set BUILD_NAME manually.

# Get the current branch name
BRANCH="$GITHUB_REF_NAME"
if [ -z "$BRANCH" ]; then
  BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null) || BRANCH=""
fi

if [ -z "$BRANCH" ]; then
  printf "uqda"
  exit 0
fi

# Remove "/" characters from the branch name if present
BRANCH=$(echo "$BRANCH" | tr -d "/")

# Default branch names: plain "uqda" (no suffix)
if [ "$BRANCH" = "master" ] || [ "$BRANCH" = "main" ]; then
  printf "uqda"
  exit 0
fi

# Any other branch: uqda-<branch> (e.g. feature/foo -> uqda-featurefoo)
printf "uqda-%s" "$BRANCH"
