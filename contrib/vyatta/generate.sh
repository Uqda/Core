#!/bin/sh

set -eu

ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

PKGARCH=${PKGARCH:-amd64}
PKGTARGET=${PKGTARGET:-vyos13}
PKGVERSION=$(sh contrib/semver/version.sh --bare)

case "$PKGTARGET" in
  edgeos2x|vyos13) ;;
  *) echo "PKGTARGET must be edgeos2x or vyos13" >&2; exit 1 ;;
esac

case "$PKGARCH" in
  amd64) GOARCH=amd64; GOMIPS= ;;
  i386) GOARCH=386; GOMIPS= ;;
  mipsel) GOARCH=mipsle; GOMIPS=softfloat ;;
  mips) GOARCH=mips64; GOMIPS= ;;
  *) echo "PKGARCH must be amd64, i386, mipsel or mips" >&2; exit 1 ;;
esac

PKGNAME="uqda-$PKGTARGET"
PKGFILE="$ROOT/$PKGNAME-$PKGVERSION-$PKGARCH.deb"
STAGE=$(mktemp -d "${TMPDIR:-/tmp}/uqda-vyatta.XXXXXX")
trap 'rm -rf "$STAGE"' EXIT HUP INT TERM
chmod 0755 "$STAGE"

GOOS=linux GOARCH="$GOARCH" GOMIPS="$GOMIPS" CGO_ENABLED=0 ./build

mkdir -p "$STAGE/usr/local/bin" "$STAGE/DEBIAN"
cp uqda uqdactl "$STAGE/usr/local/bin/"
cp -R contrib/vyatta/package/opt "$STAGE/"
cp -R contrib/vyatta/package/usr "$STAGE/"

cat > "$STAGE/DEBIAN/control" <<EOF
Package: $PKGNAME
Version: $PKGVERSION
Section: contrib/net
Priority: optional
Architecture: $PKGARCH
Maintainer: UQDA Project <https://github.com/Uqda/Core>
Depends: systemd, vyatta-cfg-system, vyatta-cfg, vyatta-op
Conflicts: vyatta-yggdrasil, yggdrasil-edgeos2x, yggdrasil-vyos13
Description: UQDA integration for $PKGTARGET
 Encrypted IPv6 mesh networking integrated with the Vyatta configuration
 and operational command interfaces.
EOF

cp contrib/vyatta/package/debian/postinst "$STAGE/DEBIAN/postinst"
cp contrib/vyatta/package/debian/prerm "$STAGE/DEBIAN/prerm"
chmod 0755 "$STAGE/DEBIAN/postinst" "$STAGE/DEBIAN/prerm"

mkdir -p "$STAGE/usr/share/doc/$PKGNAME"
cp contrib/vyatta/package/debian/copyright "$STAGE/usr/share/doc/$PKGNAME/copyright"
cp contrib/vyatta/README.md "$STAGE/usr/share/doc/$PKGNAME/README.md"

dpkg-deb --root-owner-group --build "$STAGE" "$PKGFILE"
echo "Built $PKGFILE"
