cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.0.1"
  sha256 arm:   "aa926801f0d6479458d43a5adb5a531383521fb4007ad06eb19e363706d70e79",
         intel: "6d41a7e22a6fe4a2b324788088b88dbc3a936747cc0c3e6480f1f8dd0975b5c7"
  url "https://github.com/Uqda/Core/releases/download/v#{version}/uqda-#{version}-macos-#{arch}.pkg"

  name "Uqda Core"
  desc "Encrypted IPv6 networking compatible with Yggdrasil"
  homepage "https://github.com/Uqda/Core"

  pkg "uqda-#{version}-macos-#{arch}.pkg"

  uninstall launchctl: "io.github.uqda.core",
            pkgutil:   "io.github.uqda.core"

  # The identity is deliberately retained by a normal uninstall.
  # --zap explicitly discards it and its logs.
  zap delete: ["/etc/uqda", "/Library/Logs/Uqda"]
end
