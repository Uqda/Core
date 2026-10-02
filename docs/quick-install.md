# Quick installation lifecycle

The macOS Homebrew tap and Linux quick installer use the published `v26.0.2`
release. Uqda still needs a trusted peer or a trusted local multicast network
to connect.

## Linux (systemd, x86-64 or arm64)

Ubuntu/Debian and Fedora can use the same GitHub-hosted installer. It requires
`curl`, `python3`, `tar`, `sha256sum`, and systemd. Read the script before running
it as root if your deployment policy requires review.

```sh
curl -fsSLo uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh
sudo sh uqda-install.sh install
sudo sh uqda-install.sh status
```

The installer selects the newest published release, including prereleases,
checks the archive against that release's `SHA256SUMS`, creates a persistent
identity only on first install, and starts an enabled `uqda.service` as the
unprivileged `uqda` user. It refuses to overwrite an existing service or
unmanaged binary. The admin socket stays under `/run/uqda`.

To update or remove:

```sh
curl -fsSLo uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh
sudo sh uqda-install.sh update
sudo sh uqda-install.sh uninstall
```

Normal uninstall keeps `/etc/uqda/uqda.conf` and the node's identity so a
reinstall can reconnect with the same address. To irreversibly remove identity
and configuration too, first back up any keys you need, then run:

```sh
sudo sh uqda-install.sh uninstall --purge --yes
```

The package manager remains an alternative on Debian/Ubuntu: install the
release `.deb` with `apt` or `dpkg`, upgrade with a newer `.deb`, and remove it
with `apt remove uqda`. Do not run the quick installer over an existing package
installation; it intentionally refuses to replace another service unit.

## macOS (Homebrew)

The [public Uqda Homebrew tap](https://github.com/Uqda/homebrew-tap) wraps the
native macOS `.pkg` and its launchd service. It provides `v26.0.2`:

```sh
brew install --cask Uqda/tap/uqda
brew update && brew upgrade --cask Uqda/tap/uqda
brew uninstall --cask Uqda/tap/uqda
```

Normal removal keeps `/etc/uqda/uqda.conf`. `brew uninstall --zap --cask
Uqda/tap/uqda` also removes that identity and the logs. Homebrew downloads a
separate checksum-pinned package for Intel and Apple Silicon. macOS may ask for
administrator permission because the package installs a system daemon.

## Windows

Download the x64, x86 or ARM64 `.msi` from the verified GitHub release and run
it. Installing a newer MSI upgrades the existing Uqda product line while
preserving `%ProgramData%\Uqda\uqda.conf`; remove it through Windows Installed
Apps. MSI removal also preserves the node identity. If you deliberately want
to purge that identity afterward, back it up and remove only the Uqda directory
under `%ProgramData%` manually. Native MSI install, repair and removal are
tested in the package workflow.

## Safety and compatibility

Private groups upgrading from 26.0.1 or earlier must update **every member
together**. Version 26.0.2 uses stronger Argon2id group authentication without
a legacy fallback. Keep your existing identity and strong group password;
expect a brief interruption until all members have upgraded. Public-mode
connectivity remains compatible. See [configuration](configuration.md).

Do not copy a node's private key to a second live node. Keep the configuration
and admin socket private. The Linux installer will not silently take over
another Uqda or Yggdrasil installation, and it will not claim that a release
is free of every intermittent network fault. See [security](security.md) and
[compatibility](compatibility.md).
