#!/bin/sh
set -eu
[ "$(pwd)" = "$(git rev-parse --show-toplevel)" ] || { echo "Run from the repository root" >&2; exit 1; }
PKGARCH=${PKGARCH:-amd64}
case "$PKGARCH" in
  amd64) GOARCH=amd64 ;; i386) GOARCH=386 ;; mipsel) GOARCH=mipsle ;;
  mips) GOARCH=mips ;; armhf) GOARCH=arm; export GOARM=7 ;;
  armel) GOARCH=arm; export GOARM=5 ;; arm64) GOARCH=arm64 ;;
  *) echo "Unsupported Debian architecture" >&2; exit 1 ;;
esac
export GOARCH
GOOS=linux CGO_ENABLED=0 ./build
VERSION=$(sh contrib/semver/version.sh --bare)
DEBVERSION=$(sh contrib/semver/version.sh --debian)
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT HUP INT TERM
chmod 755 "$STAGE"
mkdir -p "$STAGE/DEBIAN" "$STAGE/usr/bin" "$STAGE/usr/lib/uqda" "$STAGE/lib/systemd/system"
cp uqda uqdactl "$STAGE/usr/bin/"
cp contrib/packaging/install-config.sh "$STAGE/usr/lib/uqda/install-config"
cp contrib/systemd/uqda.service.debian "$STAGE/lib/systemd/system/uqda.service"
chmod 755 "$STAGE/usr/bin/uqda" "$STAGE/usr/bin/uqdactl" "$STAGE/usr/lib/uqda/install-config"
cat > "$STAGE/DEBIAN/control" <<EOF
Package: uqda
Version: $DEBVERSION
Section: net
Priority: optional
Architecture: $PKGARCH
Depends: systemd, passwd
Maintainer: Uqda Core maintainers
Description: Uqda Core, compatible with the Yggdrasil network
 Encrypted IPv6 networking with independently maintained Uqda tooling.
EOF
cat > "$STAGE/DEBIAN/postinst" <<'EOF'
#!/bin/sh
set -eu
if [ "${1:-}" != configure ]; then exit 0; fi
getent group uqda >/dev/null || groupadd --system uqda
install -d -m 750 -o root -g uqda /etc/uqda
UQDA_BIN=/usr/bin/uqda /usr/lib/uqda/install-config /etc/uqda/uqda.conf /etc/yggdrasil/yggdrasil.conf /etc/yggdrasil.conf
chown root:uqda /etc/uqda/uqda.conf
chmod 640 /etc/uqda/uqda.conf
systemctl daemon-reload
systemctl enable uqda
systemctl restart uqda
EOF
cat > "$STAGE/DEBIAN/prerm" <<'EOF'
#!/bin/sh
set -eu
if command -v systemctl >/dev/null; then
  systemctl stop uqda || true
  if [ "${1:-}" = remove ]; then systemctl disable uqda || true; fi
fi
# Configuration and external identity keys are operator-owned and retained.
EOF
chmod 755 "$STAGE/DEBIAN/postinst" "$STAGE/DEBIAN/prerm"
dpkg-deb --root-owner-group --build "$STAGE" "uqda-${VERSION}-${PKGARCH}.deb"
