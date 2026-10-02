cask "uqda" do
  arch arm: "arm64", intel: "amd64"
  version "26.0.2"
  sha256 arm:   "d7bc225b5ef574807c5066af6004d62a277b57f3c9f8d5114d44d9490a99c2c3",
         intel: "c38bbe60255d93f884be5fd1355dc8873fd327d9cc41fe39df95be66d225ef06"
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
