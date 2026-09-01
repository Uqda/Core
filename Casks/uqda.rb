cask "uqda" do
  version :latest
  sha256 :no_check

  url "https://github.com/Uqda/Core/releases/latest/download/install.sh"
  name "UQDA"
  desc "Encrypted, self-organizing IPv6 overlay network"
  homepage "https://github.com/Uqda/Core"

  container type: :naked

  installer script: {
    executable: "/bin/sh",
    args:       ["#{staged_path}/install.sh"],
    sudo:       true,
  }

  uninstall launchctl: "uqda",
            pkgutil:   "io.github.uqda.pkg"

  caveats <<~EOS
    UQDA's macOS package is unsigned because the project does not use a paid
    Apple Developer account. The installer verifies the package against the
    release SHA256SUMS before invoking the macOS system installer.

    Check the running node with:
      sudo uqdactl getSelf
  EOS
end
