#!/bin/sh

# Check if xar and mkbom are available
command -v xar >/dev/null 2>&1 || (
  echo "Building xar"
  sudo apt-get install libxml2-dev libssl1.0-dev zlib1g-dev -y
  mkdir -p /tmp/xar && cd /tmp/xar
  git clone https://github.com/mackyle/xar && cd xar/xar
  (sh autogen.sh && make && sudo make install) || (echo "Failed to build xar"; exit 1)
)
command -v mkbom >/dev/null 2>&1 || (
  echo "Building mkbom"
  mkdir -p /tmp/mkbom && cd /tmp/mkbom
  git clone https://github.com/hogliux/bomutils && cd bomutils
  sudo make install || (echo "Failed to build mkbom"; exit 1)
)

# Build Uqda Core
echo "running GO111MODULE=on GOOS=darwin GOARCH=${PKGARCH-amd64} ./build"
GO111MODULE=on GOOS=darwin GOARCH=${PKGARCH-amd64} ./build

# Check if we can find the files we need - they should
# exist if you are running this script from the root of
# the Uqda/Core repo and you have ran ./build
test -f uqda || (echo "uqda binary not found"; exit 1)
test -f uqdactl || (echo "uqdactl binary not found"; exit 1)
test -f contrib/macos/uqda.plist || (echo "contrib/macos/uqda.plist not found"; exit 1)
test -f contrib/semver/version.sh || (echo "contrib/semver/version.sh not found"; exit 1)

# Delete the pkgbuild folder if it already exists
test -d pkgbuild && rm -rf pkgbuild

# Create our folder structure
mkdir -p pkgbuild/scripts
mkdir -p pkgbuild/flat/base.pkg
mkdir -p pkgbuild/flat/Resources/en.lproj
mkdir -p pkgbuild/root/usr/local/bin
mkdir -p pkgbuild/root/Library/LaunchDaemons

# Copy package contents into the pkgbuild root
cp uqda pkgbuild/root/usr/local/bin
cp uqdactl pkgbuild/root/usr/local/bin
cp contrib/macos/uqda.plist pkgbuild/root/Library/LaunchDaemons

# Create the postinstall script
cat > pkgbuild/scripts/postinstall << EOF
#!/bin/sh

mkdir -p /etc/uqda

# Migrate from a previous Yggdrasil installation on this machine, if one
# is present and Uqda hasn't already been configured. This never touches
# the Yggdrasil install itself and never overwrites an existing Uqda
# config - it only helps a first-time Uqda install pick up an existing
# node identity instead of silently generating a new one.
if [ ! -f /etc/uqda/uqda.conf ] && [ -f /etc/yggdrasil.conf ];
then
  echo "Found an existing Yggdrasil configuration at /etc/yggdrasil.conf."
  echo "Validating it before migrating your node identity to Uqda..."
  if /usr/local/bin/uqda -useconffile /etc/yggdrasil.conf -checkconf;
  then
    mkdir -p /Library/Preferences/Uqda
    echo "Backing up the original to /Library/Preferences/Uqda/yggdrasil.conf.\`date +%Y%m%d\`"
    cp /etc/yggdrasil.conf /Library/Preferences/Uqda/yggdrasil.conf.\`date +%Y%m%d\`
    cp /etc/yggdrasil.conf /etc/uqda/uqda.conf
  else
    echo "The existing Yggdrasil configuration did not pass validation - not migrating it automatically."
  fi
fi

# Normalise the config if it exists, generate it if it doesn't
if [ -f /etc/uqda/uqda.conf ];
then
  mkdir -p /Library/Preferences/Uqda
  echo "Backing up configuration file to /Library/Preferences/Uqda/uqda.conf.\`date +%Y%m%d\`"
  cp /etc/uqda/uqda.conf /Library/Preferences/Uqda/uqda.conf.\`date +%Y%m%d\`
  echo "Normalising /etc/uqda/uqda.conf"
  /usr/local/bin/uqda -useconffile /Library/Preferences/Uqda/uqda.conf.\`date +%Y%m%d\` -normaliseconf > /etc/uqda/uqda.conf
else
  (umask 037 && /usr/local/bin/uqda -genconf > /etc/uqda/uqda.conf)
fi

# Unload existing Uqda launchd service, if possible
test -f /Library/LaunchDaemons/uqda.plist && (launchctl unload /Library/LaunchDaemons/uqda.plist || true)

# Load Uqda launchd service and start Uqda
launchctl load /Library/LaunchDaemons/uqda.plist
EOF

# Set execution permissions
chmod +x pkgbuild/scripts/postinstall
chmod +x pkgbuild/root/usr/local/bin/uqda
chmod +x pkgbuild/root/usr/local/bin/uqdactl

# Pack payload and scripts
( cd pkgbuild/scripts && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > pkgbuild/flat/base.pkg/Scripts
( cd pkgbuild/root && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > pkgbuild/flat/base.pkg/Payload

# Work out metadata for the package info
PKGNAME=$(sh contrib/semver/name.sh)
PKGVERSION=$(sh contrib/semver/version.sh --bare)
PKGARCH=${PKGARCH-amd64}
PAYLOADSIZE=$(( $(wc -c pkgbuild/flat/base.pkg/Payload | awk '{ print $1 }') / 1024 ))
[ "$PKGARCH" = "amd64" ] && PKGHOSTARCH="x86_64" || PKGHOSTARCH=${PKGARCH}

# Create the PackageInfo file
#
# The identifier below (io.github.uqda.core) mirrors this project's actual
# GitHub organization/repository identity (github.com/Uqda/Core) - it is
# not an invented reverse-DNS namespace claim, the same way this repo's Go
# module path is github.com/Uqda/Core.
cat > pkgbuild/flat/base.pkg/PackageInfo << EOF
<pkg-info format-version="2" identifier="io.github.uqda.core" version="${PKGVERSION}" install-location="/" auth="root">
  <payload installKBytes="${PAYLOADSIZE}" numberOfFiles="3"/>
  <scripts>
    <postinstall file="./postinstall"/>
  </scripts>
</pkg-info>
EOF

# Create the BOM
( cd pkgbuild && mkbom root flat/base.pkg/Bom )

# Create the Distribution file
cat > pkgbuild/flat/Distribution << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-script minSpecVersion="1.000000" authoringTool="com.apple.PackageMaker" authoringToolVersion="3.0.3" authoringToolBuild="174">
    <title>Uqda Core (${PKGNAME}-${PKGVERSION})</title>
    <options customize="never" allow-external-scripts="no" hostArchitectures="${PKGHOSTARCH}" />
    <domains enable_anywhere="true"/>
    <installation-check script="pm_install_check();"/>
    <script>
    function pm_install_check() {
      if(!(system.compareVersions(system.version.ProductVersion,'10.10') >= 0)) {
        my.result.title = 'Failure';
        my.result.message = 'You need at least Mac OS X 10.10 to install Uqda Core.';
        my.result.type = 'Fatal';
        return false;
      }
      return true;
    }
    </script>
    <choices-outline>
        <line choice="choice1"/>
    </choices-outline>
    <choice id="choice1" title="base">
        <pkg-ref id="io.github.uqda.core"/>
    </choice>
    <pkg-ref id="io.github.uqda.core" installKBytes="${PAYLOADSIZE}" version="${VERSION}" auth="Root">#base.pkg</pkg-ref>
</installer-script>
EOF

# Finally pack the .pkg
( cd pkgbuild/flat && xar --compression none -cf "../../${PKGNAME}-${PKGVERSION}-macos-${PKGARCH}.pkg" * )
