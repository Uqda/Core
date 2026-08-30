#!/bin/sh

# UQDA Core verified installer for Linux, macOS, FreeBSD, OpenBSD,
# EdgeOS and VyOS. It discovers the newest published GitHub release unless
# UQDA_VERSION=vX.Y.Z selects a specific one.

set -eu

REPOSITORY=${UQDA_REPOSITORY:-Uqda/Core}
VERSION=${UQDA_VERSION:-}
TEMP_BASE=${TMPDIR:-/tmp}
[ -d "$TEMP_BASE" ] || TEMP_BASE=.
DRY_RUN=0
START_SERVICE=1

say() { printf '%s\n' "[UQDA] $*"; }
die() { printf '%s\n' "[UQDA] ERROR: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"; }

usage() {
	cat <<'EOF'
Usage: install.sh [--version vX.Y.Z] [--dry-run] [--no-start]

Environment overrides:
  UQDA_VERSION           Install one specific release tag
  UQDA_TEST_OS           Override detected OS for tests
  UQDA_TEST_ARCH         Override detected CPU for tests
  UQDA_TEST_PLATFORM     systemd-deb, systemd-portable, portable, edgeos2x, or vyos13
  UQDA_RELEASE_BASE_URL  Alternate release URL (for mirrors/tests)
EOF
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--version) [ "$#" -ge 2 ] || die "--version needs a value"; VERSION=$2; shift 2 ;;
		--dry-run|--print-plan) DRY_RUN=1; shift ;;
		--no-start) START_SERVICE=0; shift ;;
		-h|--help) usage; exit 0 ;;
		--update) shift ;;
		*) die "unknown option: $1" ;;
	esac
done

if [ -z "$VERSION" ]; then
	RELEASES_FILE=${UQDA_RELEASES_FILE:-}
	if [ -z "$RELEASES_FILE" ]; then
		TMP_RELEASES=$(mktemp "$TEMP_BASE/uqda-releases.XXXXXX")
		trap 'rm -f "$TMP_RELEASES"' EXIT HUP INT TERM
		if command -v curl >/dev/null 2>&1; then
			curl --fail --location --proto '=https' --tlsv1.2 --retry 3 \
				--output "$TMP_RELEASES" "https://api.github.com/repos/$REPOSITORY/releases?per_page=20"
		elif command -v wget >/dev/null 2>&1; then
			wget --https-only --tries=3 --output-document="$TMP_RELEASES" \
				"https://api.github.com/repos/$REPOSITORY/releases?per_page=20"
		else
			die "curl or wget is required to discover the latest release"
		fi
		RELEASES_FILE=$TMP_RELEASES
	fi
	VERSION=$(sed -n 's/^[[:space:]]*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' "$RELEASES_FILE" | sed -n '1p')
fi
case "$VERSION" in
	v[0-9]*.[0-9]*.[0-9]*) ;;
	*) die "invalid release tag: $VERSION" ;;
esac
BASE_URL=${UQDA_RELEASE_BASE_URL:-https://github.com/$REPOSITORY/releases/download/$VERSION}

raw_os=${UQDA_TEST_OS:-$(uname -s)}
case "$raw_os" in
	Linux|linux) OS=linux ;;
	Darwin|darwin) OS=macos ;;
	FreeBSD|freebsd) OS=freebsd ;;
	OpenBSD|openbsd) OS=openbsd ;;
	*) die "unsupported operating system: $raw_os" ;;
esac

raw_arch=${UQDA_TEST_ARCH:-$(uname -m)}
case "$raw_arch" in
	x86_64|amd64) ARCH=amd64; DEB_ARCH=amd64 ;;
	i386|i486|i586|i686|x86) ARCH=i386; DEB_ARCH=i386 ;;
	aarch64|arm64) ARCH=arm64; DEB_ARCH=arm64 ;;
	armv7*|armhf) ARCH=armv7; DEB_ARCH=armhf ;;
	armv5*|armel) ARCH=armv5; DEB_ARCH=armel ;;
	mipsel|mipsle) ARCH=mipsel; DEB_ARCH=mipsel ;;
	mips|mips64) ARCH=mips64; DEB_ARCH=mips ;;
	*) die "unsupported CPU architecture: $raw_arch" ;;
esac

PLATFORM=${UQDA_TEST_PLATFORM:-}
if [ -z "$PLATFORM" ]; then
	if [ "$OS" = macos ]; then
		PLATFORM=launchd
	elif [ "$OS" = linux ] && { [ -d /opt/vyatta ] || [ -x /opt/vyatta/sbin/vyatta-cfg-cmd-wrapper ]; }; then
		if [ -r /etc/os-release ] && grep -Eiq '(^ID=|^NAME=).*vyos' /etc/os-release; then
			PLATFORM=vyos13
		else
			PLATFORM=edgeos2x
		fi
	elif [ "$OS" = linux ] && command -v systemctl >/dev/null 2>&1; then
		if command -v dpkg >/dev/null 2>&1; then
			PLATFORM=systemd-deb
		else
			PLATFORM=systemd-portable
		fi
	else
		PLATFORM=portable
	fi
fi

case "$PLATFORM" in
	edgeos2x|vyos13)
		[ "$OS" = linux ] || die "$PLATFORM is only valid on Linux routers"
		ASSET="uqda-$PLATFORM-${VERSION#v}-$DEB_ARCH.deb"
		METHOD=deb
		;;
	systemd-deb)
		[ "$OS" = linux ] || die "systemd package is only valid on Linux"
		ASSET="uqda-${VERSION#v}-$DEB_ARCH.deb"
		METHOD=deb
		;;
	systemd-portable)
		[ "$OS" = linux ] || die "systemd portable install is only valid on Linux"
		case "$ARCH" in
			amd64|i386|arm64|armv7|mipsel|mips64) ;;
			*) die "no portable systemd release for $ARCH" ;;
		esac
		ASSET="uqda-$VERSION-linux-$ARCH.tar.gz"
		METHOD=tar
		;;
	portable)
		case "$OS:$ARCH" in
			linux:amd64|linux:i386|linux:arm64|linux:armv7|linux:mipsel|linux:mips64|freebsd:amd64|freebsd:arm64|openbsd:amd64|openbsd:arm64) ;;
			macos:*) die "macOS requires its native installer" ;;
			*) die "no portable release for $OS/$ARCH" ;;
		esac
		ASSET="uqda-$VERSION-$OS-$ARCH.tar.gz"
		METHOD=tar
		;;
	*)
		[ "$OS" = macos ] || die "unsupported platform: $PLATFORM"
		ASSET="uqda-${VERSION#v}-macos-$ARCH.pkg"
		METHOD=pkg
		;;
esac

say "release=$VERSION os=$OS arch=$ARCH platform=$PLATFORM"
say "asset=$ASSET method=$METHOD"
[ "$DRY_RUN" -eq 0 ] || exit 0

[ "$(id -u)" -eq 0 ] || die "run this installer as root (for example: sudo sh install.sh)"
need mktemp
INSTALL_USER=${SUDO_USER:-root}
case "$INSTALL_USER" in root|'') ADMIN_GROUP=uqda ;; *) ADMIN_GROUP=$(id -gn "$INSTALL_USER" 2>/dev/null || printf 'uqda') ;; esac
case "$ADMIN_GROUP" in *[!A-Za-z0-9_.-]*) die "unsafe administration group name: $ADMIN_GROUP" ;; esac
TMPDIR_UQDA=$(mktemp -d "$TEMP_BASE/uqda-install.XXXXXX")
trap 'rm -rf "$TMPDIR_UQDA"' EXIT HUP INT TERM

download() {
	url=$1 destination=$2
	if command -v curl >/dev/null 2>&1; then
		curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$destination" "$url"
	elif command -v wget >/dev/null 2>&1; then
		wget --https-only --tries=3 --output-document="$destination" "$url"
	else
		die "curl or wget is required"
	fi
}

verify_asset() {
	file=$1 sums=$2 name=$3
	expected=$(awk -v n="$name" '$2 == n || $2 == "./" n { print $1; exit }' "$sums")
	[ -n "$expected" ] || die "$name is missing from SHA256SUMS"
	if command -v sha256sum >/dev/null 2>&1; then
		actual=$(sha256sum "$file" | awk '{print $1}')
	elif command -v shasum >/dev/null 2>&1; then
		actual=$(shasum -a 256 "$file" | awk '{print $1}')
	elif command -v openssl >/dev/null 2>&1; then
		actual=$(openssl dgst -sha256 "$file" | awk '{print $NF}')
	else
		die "no SHA-256 utility found"
	fi
	[ "$actual" = "$expected" ] || die "SHA-256 mismatch for $name"
}

download "$BASE_URL/SHA256SUMS" "$TMPDIR_UQDA/SHA256SUMS"
download "$BASE_URL/$ASSET" "$TMPDIR_UQDA/$ASSET"
verify_asset "$TMPDIR_UQDA/$ASSET" "$TMPDIR_UQDA/SHA256SUMS" "$ASSET"
say "SHA-256 verified"

install_portable_service() {
	case "$OS" in
		linux)
			mkdir -p /etc/uqda
			if [ -f /etc/uqda.conf ] && [ ! -L /etc/uqda.conf ]; then
				CONFIG_FILE=/etc/uqda.conf
			else
				CONFIG_FILE=/etc/uqda/uqda.conf
				[ -f "$CONFIG_FILE" ] || (umask 037; /usr/local/bin/uqda -genconf > "$CONFIG_FILE")
				if [ ! -e /etc/uqda.conf ] && [ ! -L /etc/uqda.conf ]; then
					ln -s "$CONFIG_FILE" /etc/uqda.conf
				fi
			fi
			if command -v systemctl >/dev/null 2>&1; then
				if [ "$ADMIN_GROUP" = uqda ] && ! getent group uqda >/dev/null 2>&1; then
					groupadd --system uqda
				fi
				chgrp "$ADMIN_GROUP" "$CONFIG_FILE"
				chmod 0640 "$CONFIG_FILE"
				cat > /etc/systemd/system/uqda.service <<'EOF'
[Unit]
Description=UQDA encrypted IPv6 mesh
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=__UQDA_ADMIN_GROUP__
ExecStart=/usr/local/bin/uqda -useconffile __UQDA_CONFIG_FILE__
Restart=on-failure
RestartSec=5s
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF
				sed -i "s/__UQDA_ADMIN_GROUP__/$ADMIN_GROUP/" /etc/systemd/system/uqda.service
				sed -i "s|__UQDA_CONFIG_FILE__|$CONFIG_FILE|" /etc/systemd/system/uqda.service
				systemctl daemon-reload
				systemctl enable uqda.service
				[ "$START_SERVICE" -eq 0 ] || systemctl restart uqda.service
			else
				say "binaries installed; no supported service manager was detected"
			fi
			;;
		freebsd)
			mkdir -p /usr/local/etc
			[ -f /usr/local/etc/uqda.conf ] || (umask 037; /usr/local/bin/uqda -genconf > /usr/local/etc/uqda.conf)
			say "binaries and configuration installed; configure the FreeBSD rc service before production use"
			;;
		openbsd)
			mkdir -p /etc
			[ -f /etc/uqda.conf ] || (umask 037; /usr/local/bin/uqda -genconf > /etc/uqda.conf)
			say "binaries and configuration installed; configure rcctl before production use"
			;;
	esac
}

case "$METHOD" in
	deb)
		need dpkg
		dpkg -i "$TMPDIR_UQDA/$ASSET"
		if [ "$INSTALL_USER" != root ] && getent group uqda >/dev/null 2>&1 && command -v usermod >/dev/null 2>&1; then
			usermod -a -G uqda "$INSTALL_USER"
		fi
		;;
	pkg)
		need installer
		installer -pkg "$TMPDIR_UQDA/$ASSET" -target /
		;;
	tar)
		need tar
		mkdir "$TMPDIR_UQDA/unpacked"
		tar -xzf "$TMPDIR_UQDA/$ASSET" -C "$TMPDIR_UQDA/unpacked" --strip-components=1
		install -m 0755 "$TMPDIR_UQDA/unpacked/uqda" /usr/local/bin/uqda
		install -m 0755 "$TMPDIR_UQDA/unpacked/uqdactl" /usr/local/bin/uqdactl
		install_portable_service
		;;
esac

verify_linux_service() {
	[ "$OS" = linux ] || return 0
	command -v systemctl >/dev/null 2>&1 || return 0
	[ "$START_SERVICE" -eq 1 ] || return 0
	if ! systemctl is-active --quiet uqda.service; then
		systemctl status --no-pager uqda.service || true
		die "uqda.service did not start"
	fi
	CTL=$(command -v uqdactl || true)
	[ -n "$CTL" ] || die "uqdactl was not installed into PATH"
	i=0
	verified=0
	while [ "$i" -lt 5 ]; do
		if [ "$INSTALL_USER" != root ] && command -v runuser >/dev/null 2>&1; then
			runuser -u "$INSTALL_USER" -- "$CTL" getSelf >/dev/null 2>&1 && verified=1 && break
		else
			"$CTL" getSelf >/dev/null 2>&1 && verified=1 && break
		fi
		sleep 1
		i=$((i + 1))
	done
	[ "$verified" -eq 1 ] || die "uqdactl cannot reach the service as $INSTALL_USER"
	say "service and non-root administration check passed"
}

verify_linux_service

say "installation completed successfully"
say "configuration was preserved if it already existed"
say "manage this node with: uqdactl getSelf or uqdactl getPeers"
