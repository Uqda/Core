cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.0-beta.1"
  sha256 arm:   "a258f5225e0697835a28a4bb50a8d93548d619e9a03560052187aad9dea2aa0a",
         intel: "d984191820e9133f9b8d8a67473e2017352e7778562b40245a0016b2a2812082"
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
