#!/bin/sh
set -eu
exec sh "$(dirname "$0")/../semver/version.sh" --installer
