#!/bin/sh

# UQDA Core cross-platform uninstaller.
# By default, it removes services and program files while preserving the
# configuration and cryptographic node identity. --purge removes those too.

set -eu

PURGE=0
ASSUME_YES=0
DRY_RUN=0
TEST_MODE=${UQDA_TEST_MODE:-0}
TEST_ROOT=${UQDA_TEST_ROOT:-}

say() { printf '%s\n' "[UQDA] $*"; }
die() { printf '%s\n' "[UQDA] ERROR: $*" >&2; exit 1; }

usage() {
	cat <<'EOF'
Usage: uninstall.sh [--purge] [--yes] [--dry-run]

Without --purge, UQDA services and program files are removed while the
configuration and cryptographic node identity are preserved for reinstalling.

Options:
  --purge    Also delete configuration, node identity, and UQDA backups
  --yes      Confirm --purge without an interactive prompt
  --dry-run  Print the removal plan without changing the system
  -h, --help Show this help
EOF
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--purge) PURGE=1 ;;
		--yes) ASSUME_YES=1 ;;
		--dry-run|--print-plan) DRY_RUN=1 ;;
		-h|--help) usage; exit 0 ;;
		*) die "unknown option: $1" ;;
	esac
	shift
done

if [ "$TEST_MODE" != 1 ] && [ "$DRY_RUN" -ne 1 ] && [ "$(id -u)" -ne 0 ]; then
	die "run as root (for example: sudo sh uninstall.sh)"
fi

if [ "$PURGE" -eq 1 ] && [ "$ASSUME_YES" -ne 1 ] && [ "$DRY_RUN" -ne 1 ]; then
	printf '%s\n' '[UQDA] WARNING: --purge permanently deletes this node identity, configuration, and backups.' >&2
	if [ ! -t 0 ]; then
		die "--purge requires --yes when input is not interactive"
	fi
	printf '%s' 'Type PURGE to continue: ' >&2
	read -r answer
	[ "$answer" = PURGE ] || die "purge cancelled"
fi

root_path() { printf '%s%s\n' "$TEST_ROOT" "$1"; }

remove_file() {
	target=$(root_path "$1")
	if [ -e "$target" ] || [ -L "$target" ]; then
		say "remove $1"
		[ "$DRY_RUN" -eq 1 ] || rm -f "$target"
	fi
}

remove_dir() {
	target=$(root_path "$1")
	if [ -d "$target" ] || [ -L "$target" ]; then
		say "remove $1"
		[ "$DRY_RUN" -eq 1 ] || rm -rf "$target"
	fi
}

run_service_command() {
	description=$1
	shift
	say "$description"
	if [ "$DRY_RUN" -ne 1 ] && [ "$TEST_MODE" != 1 ]; then
		"$@" || true
	fi
}

raw_os=${UQDA_TEST_OS:-$(uname -s)}
case "$raw_os" in
	Darwin|darwin|macos) OS=macos ;;
	Linux|linux) OS=linux ;;
	FreeBSD|freebsd) OS=freebsd ;;
	OpenBSD|openbsd) OS=openbsd ;;
	*) die "unsupported operating system: $raw_os" ;;
esac

say "uninstalling UQDA on $OS"
[ "$PURGE" -eq 0 ] || say "purge mode enabled: configuration and node identity will be deleted"

case "$OS" in
	macos)
		if [ "$TEST_MODE" != 1 ]; then
			if launchctl print system/uqda >/dev/null 2>&1; then
				run_service_command "stop launchd service uqda" launchctl bootout system/uqda
			elif [ -f /Library/LaunchDaemons/uqda.plist ]; then
				run_service_command "unload launchd service uqda" launchctl unload /Library/LaunchDaemons/uqda.plist
			fi
		fi
		remove_file /Library/LaunchDaemons/uqda.plist
		remove_file /usr/local/bin/uqda
		remove_file /usr/local/bin/uqdactl
		remove_dir /usr/local/share/uqda
		if [ "$TEST_MODE" != 1 ] && command -v pkgutil >/dev/null 2>&1 && pkgutil --pkg-info io.github.uqda.pkg >/dev/null 2>&1; then
			run_service_command "forget installer receipt io.github.uqda.pkg" pkgutil --forget io.github.uqda.pkg
		fi
		;;
	linux)
		if [ "$TEST_MODE" != 1 ] && command -v systemctl >/dev/null 2>&1; then
			for config in /config/uqda.tun[0-9]*.conf; do
				[ -f "$config" ] || continue
				interface=${config#/config/uqda.}
				interface=${interface%.conf}
				run_service_command "disable uqda@$interface.service" systemctl disable --now "uqda@$interface.service"
			done
			run_service_command "disable uqda.service" systemctl disable --now uqda.service
			run_service_command "stop uqda-default-config.service" systemctl stop uqda-default-config.service
		fi

		if [ "$TEST_MODE" != 1 ] && command -v dpkg-query >/dev/null 2>&1 && command -v dpkg >/dev/null 2>&1; then
			for package in uqda uqda-edgeos2x uqda-vyos13; do
				if dpkg-query -W -f='${Status}' "$package" 2>/dev/null | grep -q 'installed$'; then
					if [ "$PURGE" -eq 1 ]; then
						run_service_command "purge package $package" dpkg --purge "$package"
					else
						run_service_command "remove package $package" dpkg --remove "$package"
					fi
				fi
			done
		fi

		remove_file /etc/systemd/system/uqda.service
		remove_file /lib/systemd/system/uqda.service
		remove_file /lib/systemd/system/uqda-default-config.service
		remove_file /usr/lib/systemd/system/uqda.service
		remove_file /usr/lib/systemd/system/uqda-default-config.service
		remove_file /lib/systemd/system/uqda@.service
		remove_file /usr/lib/systemd/system/uqda@.service
		remove_file /usr/bin/uqda
		remove_file /usr/bin/uqdactl
		remove_file /usr/local/bin/uqda
		remove_file /usr/local/bin/uqdactl
		remove_dir /usr/share/uqda
		remove_dir /usr/local/share/uqda
		if [ "$TEST_MODE" != 1 ] && command -v systemctl >/dev/null 2>&1; then
			run_service_command "reload systemd" systemctl daemon-reload
			run_service_command "clear failed systemd state" systemctl reset-failed
		fi
		;;
	freebsd)
		if [ "$TEST_MODE" != 1 ] && command -v service >/dev/null 2>&1; then
			run_service_command "stop FreeBSD service uqda" service uqda stop
		fi
		if [ "$TEST_MODE" != 1 ] && command -v sysrc >/dev/null 2>&1; then
			run_service_command "disable FreeBSD service uqda" sysrc uqda_enable=NO
		fi
		remove_file /usr/local/etc/rc.d/uqda
		remove_file /usr/local/bin/uqda
		remove_file /usr/local/bin/uqdactl
		remove_dir /usr/local/share/uqda
		;;
	openbsd)
		if [ "$TEST_MODE" != 1 ] && command -v rcctl >/dev/null 2>&1; then
			run_service_command "stop OpenBSD service uqda" rcctl stop uqda
			run_service_command "disable OpenBSD service uqda" rcctl disable uqda
		fi
		remove_file /etc/rc.d/uqda
		remove_file /usr/local/bin/uqda
		remove_file /usr/local/bin/uqdactl
		remove_dir /usr/local/share/uqda
		;;
esac

remove_dir /var/run/uqda
remove_dir /run/uqda
remove_file /var/run/uqda.sock
remove_file /run/uqda.sock
remove_file /tmp/uqda.stdout.log
remove_file /tmp/uqda.stderr.log

if [ "$PURGE" -eq 1 ]; then
	remove_file /etc/uqda.conf
	remove_dir /etc/uqda
	remove_file /usr/local/etc/uqda.conf
	remove_dir /Library/Preferences/UQDA

	for pattern in \
		"$TEST_ROOT"/var/backups/uqda.conf.* \
		"$TEST_ROOT"/var/backups/uqda.tun[0-9]*.conf.* \
		"$TEST_ROOT"/config/uqda.tun[0-9]*.conf; do
		[ -e "$pattern" ] || [ -L "$pattern" ] || continue
		display=${pattern#"$TEST_ROOT"}
		say "remove $display"
		[ "$DRY_RUN" -eq 1 ] || rm -f "$pattern"
	done

	remove_dir /var/lib/uqda
	say "UQDA was completely removed, including configuration and node identity"
else
	say "UQDA services and program files were removed"
	say "configuration and node identity were preserved; use --purge to delete them"
fi
