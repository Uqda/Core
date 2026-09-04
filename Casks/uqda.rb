cask "uqda" do
  version "0.1.5"
  sha256 "9ed9bc866898b27584e5d1f3bf2363fd5a186a5a48b3c49f5dfbb21c280ffd2c"

  url "https://github.com/Uqda/Core/releases/download/v#{version}/install.sh"
  name "UQDA"
  desc "Encrypted, self-organizing IPv6 overlay network"
  homepage "https://github.com/Uqda/Core"

  livecheck do
    url "https://github.com/Uqda/Core/releases/latest"
    strategy :github_latest
  end

  container type: :naked

  installer script: {
    executable: "/bin/sh",
    args:       ["#{staged_path}/install.sh", "--version", "v#{version}"],
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
