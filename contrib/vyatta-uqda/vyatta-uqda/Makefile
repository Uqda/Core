#!/usr/bin/make -f
#
# vyatta-uqda - A uqda package for Ubiquiti EdgeOS, VyOS and other
# Vyatta-based systems.
# Copyright (C) 2015 Neil Alexander T.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

PKGNAME?=vyatta-$(shell cd $(BUILDDIR_CORE) && sh contrib/semver/name.sh --bare)
PKGARCH?=mipsel
PKGVERSION?=$(shell cd $(BUILDDIR_CORE) && sh contrib/semver/version.sh --bare)

# Core repository root (this Makefile lives in contrib/vyatta-uqda/vyatta-uqda/)
BUILDDIR_CORE ?= $(shell cd "$(CURDIR)/../../.." && pwd)
BRANCH ?= master

# cd $(BUILDDIR_CORE) && sh contrib/semver/version.sh | cut -c 2-
# cd $(BUILDDIR_CORE) && sh contrib/semver/name.sh

all: cross

cross: clean build-cross package cleanpackage
native: clean build-native package cleanpackage

build-cross: uqdaclone uqdabuild-cross uqdastage
build-native: uqdaclone uqdabuild-native uqdastage

package:
	@echo "Packaging $(PKGNAME) version $(PKGVERSION) for $(PKGARCH)..."
	@sed -i "s/^Architecture: .*/Architecture: $(PKGARCH)/" debian/control || \
		sed -i "" "s/^Architecture: .*/Architecture: $(PKGARCH)/" debian/control
	@sed -i "s/^Version: .*/Version: $(PKGVERSION)/" debian/control || \
		sed -i "" "s/^Version: .*/Version: $(PKGVERSION)/" debian/control
	@sed -i "s/^Package: .*/Package: $(PKGNAME)/" debian/control || \
		sed -i "" "s/^Package: .*/Package: $(PKGNAME)/" debian/control
	@cat debian/control
	tar czvf ../data.tar.gz opt usr
	cd debian && tar czvf ../../control.tar.gz *
	echo 2.0 > ../debian-binary
	ar -r ../$(PKGNAME)-$(PKGVERSION)-$(PKGARCH).deb ../debian-binary ../control.tar.gz ../data.tar.gz

cleanpackage:
	rm ../debian-binary
	rm ../control.tar.gz
	rm ../data.tar.gz

uqdaclone:
	@if [ -f "$(BUILDDIR_CORE)/go.mod" ]; then :; \
	elif [ "$(BUILDDIR_CORE)" = "/tmp/uqda" ]; then git clone -b $(BRANCH) https://github.com/Uqda/Core $(BUILDDIR_CORE); \
	else echo "uqdaclone: no $(BUILDDIR_CORE)/go.mod — set BUILDDIR_CORE to Core checkout or use /tmp/uqda for auto-clone" >&2; exit 1; fi

uqdabuild-cross:
	@[ "$(PKGARCH)" = "mipsel" ] && (cd $(BUILDDIR_CORE) && GOOS='linux' GOARCH='mipsle' GOMIPS='softfloat' ./build) || true
	@[ "$(PKGARCH)" = "mips" ] && (cd $(BUILDDIR_CORE) && GOOS='linux' GOARCH='mips64' ./build) || true
	@[ "$(PKGARCH)" = "amd64" ] && (cd $(BUILDDIR_CORE) && GOOS='linux' GOARCH='amd64' ./build) || true
	@[ "$(PKGARCH)" = "i386" ] && (cd $(BUILDDIR_CORE) && GOOS='linux' GOARCH='386' ./build) || true

uqdabuild-native:
	cd $(BUILDDIR_CORE) && GOOS='linux' ./build

uqdastage:
	mkdir -p usr/local/bin
	cp $(BUILDDIR_CORE)/uqda usr/local/bin/uqda
	cp $(BUILDDIR_CORE)/uqdactl usr/local/bin/uqdactl

clean:
	rm -rf sbin/uqda
	@[ "$(BUILDDIR_CORE)" = "/tmp/uqda" ] && rm -rf $(BUILDDIR_CORE) || true

distclean:	clean
