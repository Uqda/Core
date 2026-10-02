cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.0.4"
  sha256 arm:   "b41da5dcdcf28ac2fcca58cbc8dacba2e4b0ad583539581c76347f640ab6d720",
         intel: "674e9c634cac9bf413904367334422fcbbc97118ac3fb7cd158b3dcbbae07e2f"
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
