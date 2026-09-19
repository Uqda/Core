#!/bin/sh

# This is a lazy script to create a .deb for Debian/Ubuntu. It installs
# uqda and enables it in systemd. You can give it the PKGARCH= argument
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
PKGREPLACES=uqda

if [ $PKGBRANCH = "master" ]; then
  PKGREPLACES=uqda-develop
fi

GOLDFLAGS="-X github.com/Uqda/Core/src/config.defaultConfig=/etc/uqda/uqda.conf"
GOLDFLAGS="${GOLDFLAGS} -X github.com/Uqda/Core/src/config.defaultAdminListen=unix:///run/uqda/admin.sock"

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
Maintainer: Uqda Core maintainers
Description: Uqda Core
 Uqda Core is an independently maintained, hardened and modernized
 implementation compatible with the existing Yggdrasil network - an
 early-stage end-to-end encrypted IPv6 mesh network. It is lightweight,
 self-arranging, supported on multiple platforms and allows pretty much
 any IPv6-capable application to communicate securely with other nodes
 on the Yggdrasil network, whether they run Uqda or the original
 Yggdrasil implementation.
EOF
cat > /tmp/$PKGNAME/debian/copyright << EOF
Please see https://github.com/Uqda/Core/ and that repository's NOTICE.md
for upstream Yggdrasil and Ironwood attribution.
EOF
cat > /tmp/$PKGNAME/debian/docs << EOF
Please see https://github.com/Uqda/Core/
EOF
cat > /tmp/$PKGNAME/debian/install << EOF
usr/bin/uqda usr/bin
usr/bin/uqdactl usr/bin
lib/systemd/system/*.service lib/systemd/system
EOF
cat > /tmp/$PKGNAME/debian/postinst << EOF
#!/bin/sh

systemctl daemon-reload

if ! getent group uqda 2>&1 > /dev/null; then
  groupadd --system --force uqda
fi

if [ ! -d /etc/uqda ];
then
    mkdir -p /etc/uqda
    chown root:uqda /etc/uqda
    chmod 750 /etc/uqda
fi

# Migrate from a previous Yggdrasil installation on this machine, if one
# is present and Uqda hasn't already been configured. This never touches
# the Yggdrasil install itself (Uqda and Yggdrasil can coexist during a
# migration window) and never overwrites an existing Uqda config - it
# only helps a first-time Uqda install pick up an existing node identity
# instead of silently generating a new one.
if [ ! -f /etc/uqda/uqda.conf ] && [ -f /etc/yggdrasil/yggdrasil.conf ];
then
  echo "Found an existing Yggdrasil configuration at /etc/yggdrasil/yggdrasil.conf."
  echo "Validating it before migrating your node identity to Uqda..."
  if /usr/bin/uqda -useconffile /etc/yggdrasil/yggdrasil.conf -checkconf;
  then
    mkdir -p /var/backups
    echo "Backing up the original to /var/backups/yggdrasil.conf.\`date +%Y%m%d\`"
    cp /etc/yggdrasil/yggdrasil.conf /var/backups/yggdrasil.conf.\`date +%Y%m%d\`
    echo "Copying it to /etc/uqda/uqda.conf (your private key and identity are preserved; the original file at /etc/yggdrasil/yggdrasil.conf is left in place, untouched)"
    cp /etc/yggdrasil/yggdrasil.conf /etc/uqda/uqda.conf
    chown root:uqda /etc/uqda/uqda.conf
    chmod 640 /etc/uqda/uqda.conf
  else
    echo "The existing Yggdrasil configuration did not pass validation - not migrating it automatically."
    echo "Generating a fresh Uqda configuration instead; your old identity is still intact at /etc/yggdrasil/yggdrasil.conf if you want to migrate it by hand."
  fi
fi

if [ ! -f /etc/uqda/uqda.conf ];
then
  echo "Generating initial configuration file /etc/uqda/uqda.conf"
  (umask 037 && /usr/bin/uqda -genconf > /etc/uqda/uqda.conf)

  chown root:uqda /etc/uqda/uqda.conf
  chmod 640 /etc/uqda/uqda.conf
else
  mkdir -p /var/backups
  echo "Backing up configuration file to /var/backups/uqda.conf.\`date +%Y%m%d\`"
  cp /etc/uqda/uqda.conf /var/backups/uqda.conf.\`date +%Y%m%d\`

  echo "Normalising and updating /etc/uqda/uqda.conf"
  /usr/bin/uqda -useconf -normaliseconf < /var/backups/uqda.conf.\`date +%Y%m%d\` > /etc/uqda/uqda.conf

  chown root:uqda /etc/uqda/uqda.conf
  chmod 640 /etc/uqda/uqda.conf
fi

systemctl enable uqda
systemctl restart uqda

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

cp uqda /tmp/$PKGNAME/usr/bin/
cp uqdactl /tmp/$PKGNAME/usr/bin/
cp contrib/systemd/uqda-default-config.service.debian /tmp/$PKGNAME/lib/systemd/system/uqda-default-config.service
cp contrib/systemd/uqda.service.debian /tmp/$PKGNAME/lib/systemd/system/uqda.service

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
