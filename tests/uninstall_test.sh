#!/bin/sh

set -eu
ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TEMP_BASE=${TMPDIR:-/tmp}
[ -d "$TEMP_BASE" ] || TEMP_BASE=.
SANDBOX=$(mktemp -d "$TEMP_BASE/uqda-uninstall-test.XXXXXX")
trap 'rm -rf "$SANDBOX"' EXIT HUP INT TERM

make_file() {
	mkdir -p "$(dirname "$SANDBOX$1")"
	printf '%s\n' test > "$SANDBOX$1"
}

assert_missing() { [ ! -e "$SANDBOX$1" ] && [ ! -L "$SANDBOX$1" ] || { echo "still exists: $1" >&2; exit 1; }; }
assert_present() { [ -e "$SANDBOX$1" ] || [ -L "$SANDBOX$1" ] || { echo "missing: $1" >&2; exit 1; }; }

make_file /usr/local/bin/uqda
make_file /usr/local/bin/uqdactl
make_file /usr/local/share/uqda/uninstall.sh
make_file /Library/LaunchDaemons/uqda.plist
make_file /etc/uqda.conf
make_file /Library/Preferences/UQDA/uqda.conf.20260904
make_file /tmp/uqda.stdout.log

UQDA_TEST_MODE=1 UQDA_TEST_ROOT=$SANDBOX UQDA_TEST_OS=Darwin sh "$ROOT/uninstall.sh"
assert_missing /usr/local/bin/uqda
assert_missing /usr/local/bin/uqdactl
assert_missing /usr/local/share/uqda
assert_missing /Library/LaunchDaemons/uqda.plist
assert_missing /tmp/uqda.stdout.log
assert_present /etc/uqda.conf
assert_present /Library/Preferences/UQDA/uqda.conf.20260904

UQDA_TEST_MODE=1 UQDA_TEST_ROOT=$SANDBOX UQDA_TEST_OS=Darwin sh "$ROOT/uninstall.sh" --purge --yes
assert_missing /etc/uqda.conf
assert_missing /Library/Preferences/UQDA

make_file /usr/bin/uqda
make_file /usr/bin/uqdactl
make_file /usr/share/uqda/uninstall.sh
make_file /etc/systemd/system/uqda.service
make_file /usr/lib/systemd/system/uqda-default-config.service
make_file /etc/uqda/uqda.conf
make_file /var/backups/uqda.conf.20260904
make_file /config/uqda.tun0.conf
make_file /var/backups/uqda.tun0.conf.20260904
make_file /run/uqda/uqda.sock

UQDA_TEST_MODE=1 UQDA_TEST_ROOT=$SANDBOX UQDA_TEST_OS=Linux sh "$ROOT/uninstall.sh"
assert_missing /usr/bin/uqda
assert_missing /usr/bin/uqdactl
assert_missing /usr/share/uqda
assert_missing /etc/systemd/system/uqda.service
assert_missing /usr/lib/systemd/system/uqda-default-config.service
assert_missing /run/uqda
assert_present /etc/uqda/uqda.conf
assert_present /var/backups/uqda.conf.20260904
assert_present /config/uqda.tun0.conf

before=$(find "$SANDBOX" -print | sort)
dry_output=$(UQDA_TEST_MODE=1 UQDA_TEST_ROOT=$SANDBOX UQDA_TEST_OS=Linux sh "$ROOT/uninstall.sh" --purge --dry-run)
after=$(find "$SANDBOX" -print | sort)
[ "$before" = "$after" ] || { echo "dry-run changed files" >&2; exit 1; }
printf '%s\n' "$dry_output" | grep -F "dry-run mode enabled: no changes will be made" >/dev/null
printf '%s\n' "$dry_output" | grep -F "would remove /etc/uqda" >/dev/null
printf '%s\n' "$dry_output" | grep -F "dry run completed; no files were changed" >/dev/null
if printf '%s\n' "$dry_output" | grep -F "was completely removed" >/dev/null; then
	echo "dry-run falsely reported completed removal" >&2
	exit 1
fi

UQDA_TEST_MODE=1 UQDA_TEST_ROOT=$SANDBOX UQDA_TEST_OS=Linux sh "$ROOT/uninstall.sh" --purge --yes
assert_missing /etc/uqda
assert_missing /var/backups/uqda.conf.20260904
assert_missing /config/uqda.tun0.conf
assert_missing /var/backups/uqda.tun0.conf.20260904

if UQDA_TEST_MODE=1 UQDA_TEST_ROOT=$SANDBOX UQDA_TEST_OS=Linux sh "$ROOT/uninstall.sh" --purge </dev/null >/dev/null 2>&1; then
	echo "non-interactive purge without --yes was accepted" >&2
	exit 1
fi

echo "uninstaller preservation, purge, and dry-run tests passed"
