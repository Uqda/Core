#!/bin/sh
set -eu
command -v pkgbuild >/dev/null
command -v productbuild >/dev/null
PKGARCH=${PKGARCH:-arm64}
GOOS=darwin GOARCH="$PKGARCH" ./build
VERSION=$(sh contrib/semver/version.sh --bare)
INSTALLVERSION=$(sh contrib/semver/version.sh --installer)
TITLE=$(sh contrib/semver/version.sh --display)
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT HUP INT TERM
mkdir -p "$STAGE/root/usr/local/bin" "$STAGE/root/Library/LaunchDaemons" "$STAGE/scripts"
cp uqda uqdactl "$STAGE/root/usr/local/bin/"
cp contrib/macos/uqda.plist "$STAGE/root/Library/LaunchDaemons/"
cp contrib/packaging/install-config.sh "$STAGE/scripts/install-config"
cat > "$STAGE/scripts/postinstall" <<'EOF'
#!/bin/sh
set -eu
umask 077
UQDA_BIN=/usr/local/bin/uqda sh "$(dirname "$0")/install-config" /etc/uqda/uqda.conf /etc/yggdrasil.conf /etc/yggdrasil/yggdrasil.conf
chmod 600 /etc/uqda/uqda.conf
mkdir -p /var/run/uqda /Library/Logs/Uqda
launchctl bootout system /Library/LaunchDaemons/uqda.plist 2>/dev/null || true
launchctl bootstrap system /Library/LaunchDaemons/uqda.plist
EOF
chmod 755 "$STAGE/scripts/postinstall" "$STAGE/scripts/install-config" "$STAGE/root/usr/local/bin/uqda" "$STAGE/root/usr/local/bin/uqdactl"
pkgbuild --root "$STAGE/root" --scripts "$STAGE/scripts" --identifier io.github.uqda.core --version "$INSTALLVERSION" --install-location / "$STAGE/core.pkg"
cat > "$STAGE/distribution.xml" <<EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="2">
  <title>$TITLE</title>
  <options customize="never" require-scripts="true"/>
  <domains enable_localSystem="true"/>
  <choices-outline><line choice="core"/></choices-outline>
  <choice id="core" visible="false"><pkg-ref id="io.github.uqda.core"/></choice>
  <pkg-ref id="io.github.uqda.core" version="$INSTALLVERSION">core.pkg</pkg-ref>
</installer-gui-script>
EOF
productbuild --distribution "$STAGE/distribution.xml" --package-path "$STAGE" "uqda-${VERSION}-macos-${PKGARCH}.pkg"
