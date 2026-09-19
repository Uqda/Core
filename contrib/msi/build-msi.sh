#!/bin/sh

# This script generates an MSI file for Uqda Core for a given architecture. It
# needs to run on Windows within MSYS2 and Go 1.25 or later must be installed on
# the system and within the PATH. This is ran currently by GitHub Actions (see
# the workflows in the repository).
#
# Originally authored for Yggdrasil by Neil Alexander
# <neilalexander@users.noreply.github.com>; adapted for Uqda Core.

set -eu

# Get arch from command line if given
PKGARCH=${1:-}
if [ "${PKGARCH}" = "" ];
then
  echo "tell me the architecture: x86, x64 or arm64"
  exit 1
fi

# WiX v3 must be installed by the build environment.
command -v candle >/dev/null
command -v light >/dev/null

case "$PKGARCH" in
  x64) GOARCH=amd64 ;; x86) GOARCH=386 ;; arm64) GOARCH=arm64 ;;
  *) echo "Unsupported MSI architecture" >&2; exit 1 ;;
esac
export GOARCH
GOOS=windows CGO_ENABLED=0 ./build

# Create the postinstall script. This also migrates an existing Yggdrasil
# configuration on this machine if present and Uqda hasn't already been
# configured - it never touches the Yggdrasil install itself, and never
# overwrites an existing Uqda config.
cp contrib/msi/updateconfig.bat updateconfig.bat

# Work out metadata for the package info
PKGNAME=$(sh contrib/semver/name.sh)
PKGVERSION=$(sh contrib/semver/version.sh --bare)
PKGVERSIONMS=$(sh contrib/msi/msversion.sh)
([ "${PKGARCH}" = "x64" ] || [ "${PKGARCH}" = "arm64" ]) && \
  PKGGUID="f1764223-fadb-499a-99d8-0e1bb119c1f5" PKGINSTFOLDER="ProgramFiles64Folder" || \
  PKGGUID="28b35855-6799-429f-9b77-1f4eb26c8dc8" PKGINSTFOLDER="ProgramFilesFolder"

# Download the Wintun driver
if [ ! -d wintun ];
then
  curl -o wintun.zip https://www.wintun.net/builds/wintun-0.14.1.zip
  if [ `sha256sum wintun.zip | cut -f 1 -d " "` != "07c256185d6ee3652e09fa55c0b673e2624b565e02c4b9091c79ca7d2f24ef51" ];
  then
    echo "wintun package didn't match expected checksum"
    exit 1
  fi
  unzip wintun.zip
fi
if [ $PKGARCH = "x64" ]; then
  PKGWINTUNDLL=wintun/bin/amd64/wintun.dll
elif [ $PKGARCH = "x86" ]; then
  PKGWINTUNDLL=wintun/bin/x86/wintun.dll
elif [ $PKGARCH = "arm64" ]; then
  PKGWINTUNDLL=wintun/bin/arm64/wintun.dll
else
  echo "wasn't sure which architecture to get wintun for"
  exit 1
fi

PKGDISPLAYNAME=$(sh contrib/semver/version.sh --display)

# Generate the wix.xml file
#
# UpgradeCode and the Component Guids below are freshly generated, not
# carried over from the old Yggdrasil installer: Uqda is a distinct
# product from Windows Installer's point of view (it can be installed
# alongside an existing Yggdrasil install during a migration window,
# per updateconfig.bat above), not an in-place upgrade of it, so treating
# it as the same product line via a shared UpgradeCode would be wrong.
cat > wix.xml << EOF
<?xml version="1.0" encoding="windows-1252"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product
    Name="${PKGDISPLAYNAME}"
    Id="*"
    UpgradeCode="${PKGGUID}"
    Language="1033"
    Codepage="1252"
    Version="${PKGVERSIONMS}"
    Manufacturer="github.com/Uqda">

    <Package
      Id="*"
      Keywords="Installer"
      Description="Uqda Core Installer"
      Comments="Uqda Core standalone router for Windows - an independently maintained, hardened and modernized implementation compatible with the existing Yggdrasil network."
      Manufacturer="github.com/Uqda"
      InstallerVersion="500"
      InstallScope="perMachine"
      Languages="1033"
      Compressed="yes"
      SummaryCodepage="1252" />

    <MajorUpgrade
      DowngradeErrorMessage="A newer Uqda Core version is already installed." />

    <Media
      Id="1"
      Cabinet="Media.cab"
      EmbedCab="yes"
      CompressionLevel="high" />

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="${PKGINSTFOLDER}" Name="PFiles">
        <Directory Id="UqdaInstallFolder" Name="Uqda">

          <Component Id="MainExecutable" Guid="160e04b8-7cd1-4302-8dd3-7703b91e566e">
            <File
              Id="Uqda"
              Name="uqda.exe"
              DiskId="1"
              Source="uqda.exe"
              KeyPath="yes" />

            <File
              Id="Wintun"
              Name="wintun.dll"
              DiskId="1"
              Source="${PKGWINTUNDLL}" />

            <ServiceInstall
              Id="ServiceInstaller"
              Account="LocalSystem"
              Description="Uqda Core router process"
              DisplayName="Uqda Service"
              ErrorControl="normal"
              LoadOrderGroup="NetworkProvider"
              Name="Uqda"
              Start="auto"
              Type="ownProcess"
              Arguments='-useconffile "[CommonAppDataFolder]Uqda\\uqda.conf" -logto "[CommonAppDataFolder]Uqda\\uqda.log"'
              Vital="yes" />

            <ServiceControl
              Id="ServiceControl"
              Name="uqda"
              Start="install"
              Stop="both"
              Remove="uninstall" />
          </Component>

          <Component Id="CtrlExecutable" Guid="f3d39067-7381-4dc9-921b-9ee9be0ab3ac">
            <File
              Id="Uqdactl"
              Name="uqdactl.exe"
              DiskId="1"
              Source="uqdactl.exe"
              KeyPath="yes"/>
          </Component>

          <Component Id="ConfigScript" Guid="145c5c70-2cc5-45c5-adc3-0a9b4dc2cdf5">
            <File
              Id="Configbat"
              Name="updateconfig.bat"
              DiskId="1"
              Source="updateconfig.bat"
              KeyPath="yes"/>
          </Component>
        </Directory>
      </Directory>
    </Directory>

    <Feature Id="UqdaFeature" Title="Uqda" Level="1">
      <ComponentRef Id="MainExecutable" />
      <ComponentRef Id="CtrlExecutable" />
      <ComponentRef Id="ConfigScript" />
    </Feature>

    <CustomAction
      Id="UpdateGenerateConfig"
      Directory="UqdaInstallFolder"
      ExeCommand="cmd.exe /c updateconfig.bat"
      Execute="deferred"
      Return="check"
      Impersonate="no" />

    <InstallExecuteSequence>
      <Custom
        Action="UpdateGenerateConfig"
        Before="StartServices">
          NOT Installed AND NOT REMOVE
      </Custom>
    </InstallExecuteSequence>

  </Product>
</Wix>
EOF

# Generate the MSI
CANDLEFLAGS="-nologo"
LIGHTFLAGS="-nologo -spdb -sice:ICE71 -sice:ICE61"
candle $CANDLEFLAGS -out ${PKGNAME}-${PKGVERSION}-${PKGARCH}.wixobj -arch ${PKGARCH} wix.xml && \
light $LIGHTFLAGS -ext WixUtilExtension.dll -out ${PKGNAME}-${PKGVERSION}-${PKGARCH}.msi ${PKGNAME}-${PKGVERSION}-${PKGARCH}.wixobj
