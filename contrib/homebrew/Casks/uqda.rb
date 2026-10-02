cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.0.3"
  sha256 arm:   "33ed061f30bc943ca6325aa11afb83745101d3a3d254428b13ab96e97d7d7aa9",
         intel: "1a78ee05671a04146c27c58001332153f219d87b4c7c3ded5c0c8554ef611509"
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
