#!/bin/sh

# This is a lazy script to create a .deb for Debian/Ubuntu. It installs
# Uqda and enables it in systemd. You can give it the PKGARCH= argument
# i.e. PKGARCH=i386 sh contrib/deb/generate.sh

if [ `pwd` != `git rev-parse --show-toplevel` ]
then
  echo "You should run this script from the top-level directory of the git repo"
  exit 1
fi

PKGBRANCH=$(basename `git name-rev --name-only HEAD`)
PKGNAME=$(sh contrib/semver/name.sh)
PKGVERSION=$(sh contrib/semver/version.sh --bare)
PKGARCH=${PKGARCH-amd64}
PKGFILE=$PKGNAME-$PKGVERSION-$PKGARCH.deb
PKGREPLACES=Uqda

if [ $PKGBRANCH = "master" ]; then
  PKGREPLACES=Uqda-develop
fi

GOLDFLAGS="-X github.com/Uqda/Core/src/config.defaultConfig=/etc/uqda/uqda.conf"
GOLDFLAGS="${GOLDFLAGS} -X github.com/Uqda/Core/src/config.defaultAdminListen=unix:///var/run/uqda/uqda.sock"

if [ $PKGARCH = "amd64" ]; then GOARCH=amd64 GOOS=linux ./build -l "${GOLDFLAGS}"
elif [ $PKGARCH = "i386" ]; then GOARCH=386 GOOS=linux ./build -l "${GOLDFLAGS}"
elif [ $PKGARCH = "mipsel" ]; then GOARCH=mipsle GOOS=linux ./build -l "${GOLDFLAGS}"
elif [ $PKGARCH = "mips" ]; then GOARCH=mips64 GOOS=linux ./build -l "${GOLDFLAGS}"
elif [ $PKGARCH = "armhf" ]; then GOARCH=arm GOOS=linux GOARM=6 ./build -l "${GOLDFLAGS}"
elif [ $PKGARCH = "arm64" ]; then GOARCH=arm64 GOOS=linux ./build -l "${GOLDFLAGS}"
elif [ $PKGARCH = "armel" ]; then GOARCH=arm GOOS=linux GOARM=5 ./build -l "${GOLDFLAGS}"
else
  echo "Specify PKGARCH=amd64,i386,mips,mipsel,armhf,arm64,armel"
  exit 1
fi

echo "Building $PKGFILE"

mkdir -p /tmp/$PKGNAME/
mkdir -p /tmp/$PKGNAME/debian/
mkdir -p /tmp/$PKGNAME/usr/bin/
mkdir -p /tmp/$PKGNAME/lib/systemd/system/

cat > /tmp/$PKGNAME/debian/changelog << EOF
Please see https://github.com/Uqda/Core/
EOF
echo 9 > /tmp/$PKGNAME/debian/compat
cat > /tmp/$PKGNAME/debian/control << EOF
Package: $PKGNAME
Version: $PKGVERSION
Section: golang
Priority: optional
Architecture: $PKGARCH
Replaces: $PKGREPLACES
Conflicts: $PKGREPLACES
Depends: systemd
Maintainer: Neil Alexander <neilalexander@users.noreply.github.com>
Description: Uqda Network
 Uqda is an early-stage implementation of a fully end-to-end encrypted IPv6
 network. It is lightweight, self-arranging, supported on multiple platforms and
 allows pretty much any IPv6-capable application to communicate securely with
 other Uqda nodes.
EOF
cat > /tmp/$PKGNAME/debian/copyright << EOF
Please see https://github.com/Uqda/Core/
EOF
cat > /tmp/$PKGNAME/debian/docs << EOF
Please see https://github.com/Uqda/Core/
EOF
cat > /tmp/$PKGNAME/debian/install << EOF
usr/bin/uqda usr/bin
usr/bin/uqdactl usr/bin
lib/systemd/system/uqda.service lib/systemd/system
lib/systemd/system/uqda-default-config.service lib/systemd/system
EOF
cat > /tmp/$PKGNAME/debian/postinst << EOF
#!/bin/sh

set -e

# Reload systemd to recognize new service files
systemctl daemon-reload

# Create uqda group if it doesn't exist
if ! getent group uqda >/dev/null 2>&1; then
  groupadd --system uqda || true
fi

# Create configuration directory
if [ ! -d /etc/uqda ]; then
  mkdir -p /etc/uqda
  chown root:uqda /etc/uqda
  chmod 750 /etc/uqda
fi

# Check if binaries are installed
if [ ! -f /usr/bin/uqda ]; then
  echo "Error: /usr/bin/uqda not found after installation"
  exit 1
fi

# Handle configuration file
if [ -f /etc/uqda.conf ]; then
  # Move old config file if it exists in root
  mv /etc/uqda.conf /etc/uqda/uqda.conf
fi

if [ -f /etc/uqda/uqda.conf ]; then
  # Backup existing configuration
  mkdir -p /var/backups
  BACKUP_FILE="/var/backups/uqda.conf.\$(date +%Y%m%d)"
  echo "Backing up configuration file to \$BACKUP_FILE"
  cp /etc/uqda/uqda.conf "\$BACKUP_FILE"

  # Normalize and update configuration
  echo "Normalising and updating /etc/uqda/uqda.conf"
  /usr/bin/uqda -useconffile "\$BACKUP_FILE" -normaliseconf > /etc/uqda/uqda.conf.new
  mv /etc/uqda/uqda.conf.new /etc/uqda/uqda.conf

  chown root:uqda /etc/uqda/uqda.conf
  chmod 640 /etc/uqda/uqda.conf
else
  # Generate initial configuration
  echo "Generating initial configuration file /etc/uqda/uqda.conf"
  /usr/bin/uqda -genconf > /etc/uqda/uqda.conf

  chown root:uqda /etc/uqda/uqda.conf
  chmod 640 /etc/uqda/uqda.conf
fi

# Enable and start the service
echo "Enabling uqda service..."
systemctl enable uqda.service || true
systemctl enable uqda-default-config.service || true

echo "Starting uqda service..."
systemctl start uqda.service || true

echo ""
echo "Uqda Network has been installed and started successfully!"
echo "Your node information:"
/usr/bin/uqdactl getSelf 2>/dev/null || echo "  (Service is starting, run 'sudo uqdactl getSelf' in a moment)"
echo ""
echo "To add peers, use: sudo uqdactl addPeer tcp://peer.example.com:12345"
echo "To view peers, use: sudo uqdactl getPeers"

exit 0
EOF
cat > /tmp/$PKGNAME/debian/prerm << EOF
#!/bin/sh
if command -v systemctl >/dev/null; then
  if systemctl is-active --quiet uqda; then
    systemctl stop uqda || true
  fi
  systemctl disable uqda || true
fi
EOF

# Check if binaries exist before copying
if [ ! -f uqda ]; then
  echo "Error: uqda binary not found. Make sure build completed successfully."
  exit 1
fi
if [ ! -f uqdactl ]; then
  echo "Error: uqdactl binary not found. Make sure build completed successfully."
  exit 1
fi

# Check if systemd service files exist
if [ ! -f contrib/systemd/uqda.service.debian ]; then
  echo "Error: contrib/systemd/uqda.service.debian not found."
  exit 1
fi
if [ ! -f contrib/systemd/uqda-default-config.service.debian ]; then
  echo "Error: contrib/systemd/uqda-default-config.service.debian not found."
  exit 1
fi

cp uqda /tmp/$PKGNAME/usr/bin/
cp uqdactl /tmp/$PKGNAME/usr/bin/
cp contrib/systemd/uqda-default-config.service.debian /tmp/$PKGNAME/lib/systemd/system/uqda-default-config.service
cp contrib/systemd/uqda.service.debian /tmp/$PKGNAME/lib/systemd/system/uqda.service

# Verify files were copied
if [ ! -f /tmp/$PKGNAME/usr/bin/uqda ]; then
  echo "Error: Failed to copy uqda binary"
  exit 1
fi
if [ ! -f /tmp/$PKGNAME/usr/bin/uqdactl ]; then
  echo "Error: Failed to copy uqdactl binary"
  exit 1
fi
if [ ! -f /tmp/$PKGNAME/lib/systemd/system/uqda.service ]; then
  echo "Error: Failed to copy uqda.service"
  exit 1
fi

tar --no-xattrs -czvf /tmp/$PKGNAME/data.tar.gz -C /tmp/$PKGNAME/ \
  usr/bin/uqda usr/bin/uqdactl \
  lib/systemd/system/uqda.service \
  lib/systemd/system/uqda-default-config.service
tar --no-xattrs -czvf /tmp/$PKGNAME/control.tar.gz -C /tmp/$PKGNAME/debian .
echo 2.0 > /tmp/$PKGNAME/debian-binary

ar -r $PKGFILE \
  /tmp/$PKGNAME/debian-binary \
  /tmp/$PKGNAME/control.tar.gz \
  /tmp/$PKGNAME/data.tar.gz

rm -rf /tmp/$PKGNAME
