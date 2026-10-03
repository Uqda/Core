cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.0.5"
  sha256 arm:   "a33059c14babd23baa4f24aa0c907d47d2f1eeeacab792707f880880d32cc2e9",
         intel: "c86078e58498353e5858bc3008a38642e502dd22c5f9a72caa3a7a20cccac1a2"
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
