cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.1.0"
  sha256 arm:   "434faf1b782865d587e4e6337cfc1faa52e7956415843317555625c92d8434d9",
         intel: "62d31ed664e78305cbf7e5a64402ff2946a8d6a6d3ee40eace3b36d39aab64b4"
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
