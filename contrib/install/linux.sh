#!/bin/sh
# Uqda Core quick install/update/uninstall for systemd Linux (amd64, arm64).
set -eu
umask 077

usage() {
  echo 'Usage: sudo sh linux.sh install|update|status|uninstall [--purge --yes]' >&2
  exit 2
}
die() { echo "uqda-install: $*" >&2; exit 1; }
say() { echo "uqda-install: $*"; }

ACTION=${1:-}
case "$ACTION" in install|update|status|uninstall) ;; *) usage ;; esac
[ "$(id -u)" -eq 0 ] || die 'run as root (sudo)'
[ "$(uname -s)" = Linux ] || die 'Linux only'
command -v systemctl >/dev/null 2>&1 || die 'systemd is required'

ROOT=/opt/uqda
CURRENT=$ROOT/current
SERVICE=/etc/systemd/system/uqda.service
CONFIG=/etc/uqda/uqda.conf
BIN=/usr/local/bin/uqda
CTL=/usr/local/bin/uqdactl
MARKER='# Managed by Uqda quick installer. Do not edit; use systemd drop-ins.'

managed() { [ -f "$SERVICE" ] && grep -Fqx "$MARKER" "$SERVICE"; }
check_ownership() {
  if [ -e "$SERVICE" ] && ! managed; then die "$SERVICE exists and is not managed by this installer"; fi
  for link in "$BIN" "$CTL"; do
    if [ -e "$link" ] || [ -L "$link" ]; then
      [ -L "$link" ] || die "$link exists and is not an installer symlink"
      case "$link:$(readlink "$link")" in
        "$BIN:$CURRENT/uqda"|"$CTL:$CURRENT/uqdactl") ;;
        *) die "$link points outside its managed target" ;;
      esac
    fi
  done
  if [ -e "$ROOT" ] && [ ! -d "$ROOT/releases" ]; then die "$ROOT is not a managed installation"; fi
  FRAGMENT=$(systemctl show -p FragmentPath --value uqda 2>/dev/null || true)
  case "$FRAGMENT" in ''|"$SERVICE") ;; *) die "existing Uqda service at $FRAGMENT is not managed by this installer" ;; esac
}

if [ "$ACTION" = status ]; then
  if [ -L "$CURRENT" ]; then
    say "version: $(basename "$(readlink "$CURRENT")")"
  elif [ -e "$SERVICE" ]; then
    say 'existing service is not managed by the quick installer'
  else
    say 'not installed'
  fi
  systemctl is-enabled uqda 2>/dev/null || true
  systemctl is-active uqda 2>/dev/null || true
  exit 0
fi

check_ownership
if [ "$ACTION" = uninstall ]; then
  PURGE=no
  YES=no
  shift
  for option in "$@"; do
    case "$option" in --purge) PURGE=yes ;; --yes) YES=yes ;; *) usage ;; esac
  done
  [ "$PURGE" = no ] || [ "$YES" = yes ] || die '--purge deletes the node identity; repeat with --purge --yes'
  CREATED_USER=no
  CREATED_GROUP=no
  [ ! -f /etc/uqda/.created-user ] || CREATED_USER=yes
  [ ! -f /etc/uqda/.created-group ] || CREATED_GROUP=yes
  [ -f "$SERVICE" ] && { systemctl disable --now uqda || true; rm -f -- "$SERVICE"; systemctl daemon-reload; }
  for link in "$BIN" "$CTL"; do [ ! -L "$link" ] || rm -f -- "$link"; done
  if [ -d "$ROOT/releases" ]; then rm -r -- "$ROOT"; fi
  if [ "$PURGE" = yes ]; then
    [ ! -d /etc/uqda ] || rm -r -- /etc/uqda
    if [ "$CREATED_USER" = yes ] && id -u uqda >/dev/null 2>&1; then userdel uqda; fi
    if [ "$CREATED_GROUP" = yes ] && getent group uqda >/dev/null 2>&1; then groupdel uqda; fi
    say 'removed program, service, and identity'
  else
    say "removed program and service; identity kept at $CONFIG"
  fi
  exit 0
fi

if [ "$ACTION" = update ] && [ ! -L "$CURRENT" ]; then die 'not installed; use install'; fi
for tool in curl python3 sha256sum tar install; do command -v "$tool" >/dev/null 2>&1 || die "$tool is required"; done
ARCH=$(uname -m)
case "$ARCH" in x86_64|amd64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; *) die "unsupported architecture: $ARCH" ;; esac

# The GitHub latest-release endpoint omits prereleases. Select the newest
# published release, including beta releases, from the ordered releases API.
VERSION=${UQDA_VERSION:-}
if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL --retry 3 'https://api.github.com/repos/Uqda/Core/releases?per_page=30' |
    python3 -c 'import json,sys; print(next((r["tag_name"] for r in json.load(sys.stdin) if not r["draft"]), ""))')
fi
case "$VERSION" in v[0-9]* ) ;; *) die "invalid release tag: $VERSION" ;; esac
case "$VERSION" in *[!a-zA-Z0-9._-]* ) die "invalid release tag: $VERSION" ;; esac
BARE=${VERSION#v}
ASSET="uqda-${BARE}-linux-${ARCH}.tar.gz"
URL=${UQDA_DOWNLOAD_BASE_URL:-"https://github.com/Uqda/Core/releases/download/$VERSION"}
case "$URL" in
  https://github.com/Uqda/Core/releases/download/*) ;;
  file://*) [ "${UQDA_INSTALL_TEST_MODE:-}" = 1 ] || die 'file download override is test-only' ;;
  *) die 'download base URL must be the official GitHub release' ;;
esac
TEMP=$(mktemp -d)
trap 'rm -r -- "$TEMP"' EXIT HUP INT TERM
say "downloading $ASSET"
curl -fsSL --retry 3 "$URL/SHA256SUMS" -o "$TEMP/SHA256SUMS"
curl -fsSL --retry 3 "$URL/$ASSET" -o "$TEMP/$ASSET"
EXPECTED=$(awk -v name="$ASSET" '$2 == name {print $1}' "$TEMP/SHA256SUMS")
case "$EXPECTED" in *[!a-fA-F0-9]*|'') die "missing or invalid SHA-256 for $ASSET" ;; esac
[ "${#EXPECTED}" -eq 64 ] || die "invalid SHA-256 length for $ASSET"
(cd "$TEMP" && printf '%s  %s\n' "$EXPECTED" "$ASSET" | sha256sum -c -) || die 'asset checksum mismatch'

mkdir "$TEMP/payload"
tar -tzf "$TEMP/$ASSET" | while IFS= read -r entry; do
  case "$entry" in uqda|uqdactl|uqda.service|install-config.sh) ;; *) die "unexpected archive entry: $entry" ;; esac
done
tar -C "$TEMP/payload" -xzf "$TEMP/$ASSET" --no-same-owner
for file in uqda uqdactl uqda.service install-config.sh; do [ -f "$TEMP/payload/$file" ] || die "missing $file"; done
grep -Fqx "$MARKER" "$TEMP/payload/uqda.service" || die 'service unit is not the quick-install unit'
"$TEMP/payload/uqda" -version >/dev/null 2>&1 || die 'downloaded binary cannot execute'

CREATED_GROUP=no
CREATED_USER=no
if ! getent group uqda >/dev/null 2>&1; then groupadd --system uqda; CREATED_GROUP=yes; fi
if ! id -u uqda >/dev/null 2>&1; then useradd --system -g uqda -M -s /usr/sbin/nologin uqda; CREATED_USER=yes; fi
install -d -m 750 -o root -g uqda /etc/uqda
[ "$CREATED_USER" = no ] || touch /etc/uqda/.created-user
[ "$CREATED_GROUP" = no ] || touch /etc/uqda/.created-group
UQDA_BIN="$TEMP/payload/uqda" sh "$TEMP/payload/install-config.sh" "$CONFIG" \
  /etc/yggdrasil/yggdrasil.conf /etc/yggdrasil.conf
chown uqda:uqda "$CONFIG"
chmod 600 "$CONFIG"

install -d -m 755 -o root -g root "$ROOT" "$ROOT/releases"
TARGET="$ROOT/releases/$VERSION"
if [ ! -d "$TARGET" ]; then
  install -d -m 755 -o root -g root "$TARGET"
  install -m 755 -o root -g root "$TEMP/payload/uqda" "$TEMP/payload/uqdactl" "$TARGET/"
else
  [ -x "$TARGET/uqda" ] && [ -x "$TARGET/uqdactl" ] || die "incomplete existing release at $TARGET"
  cmp -s "$TEMP/payload/uqda" "$TARGET/uqda" && cmp -s "$TEMP/payload/uqdactl" "$TARGET/uqdactl" || die "existing release $VERSION differs from download"
fi
OLD=
[ ! -L "$CURRENT" ] || OLD=$(readlink "$CURRENT")
ln -s "releases/$VERSION" "$ROOT/.current.$$"
mv -Tf -- "$ROOT/.current.$$" "$CURRENT"
if [ ! -e "$SERVICE" ]; then install -m 644 -o root -g root "$TEMP/payload/uqda.service" "$SERVICE"; fi
[ -L "$BIN" ] || ln -s "$CURRENT/uqda" "$BIN"
[ -L "$CTL" ] || ln -s "$CURRENT/uqdactl" "$CTL"
systemctl daemon-reload
systemctl enable uqda >/dev/null
if ! systemctl restart uqda; then
  if [ -n "$OLD" ]; then
    ln -s "$OLD" "$ROOT/.rollback.$$"
    mv -Tf -- "$ROOT/.rollback.$$" "$CURRENT"
    systemctl restart uqda || true
  fi
  die 'service failed to start; prior release restored when available'
fi
sleep 2
if ! systemctl is-active --quiet uqda; then
  if [ -n "$OLD" ]; then
    ln -s "$OLD" "$ROOT/.rollback.$$"
    mv -Tf -- "$ROOT/.rollback.$$" "$CURRENT"
    systemctl restart uqda || true
  fi
  die 'service stopped after startup; prior release restored when available'
fi
say "running $VERSION; identity preserved at $CONFIG"
