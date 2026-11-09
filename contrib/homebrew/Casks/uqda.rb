cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.1.1"
  sha256 arm:   "abf3216785c7d6be435ed932e292e6c2171f3cc175ce2a437bea8fa5b100c8c3",
         intel: "7f817b4c1a4843b2eee347cd8e8810a25642c629afda4cedb3df2fb01a25f842"
  url "https://github.com/Uqda/Core/releases/download/v#{version}/uqda-#{version}-macos-#{arch}.pkg"

  name "Uqda Core"
  desc "Encrypted IPv6 networking compatible with Yggdrasil"
  homepage "https://github.com/Uqda/Core"

  pkg "uqda-#{version}-macos-#{arch}.pkg"

  uninstall launchctl: "io.github.uqda.core",
            pkgutil:   "io.github.uqda.core"

  # A normal uninstall retains the node identity.
  zap delete: ["/etc/uqda", "/Library/Logs/Uqda"]
end
