# Uqda Core 26.0.3

A new patch release with complete command discovery and actionable local
permission diagnostics. Published 26.0.2 artifacts and its tag remain unchanged.

## Changes

- `uqda help` documents every built-in admin operation, daily aliases, daemon
  flags, examples, security boundaries, exit codes and installation lifecycle.
- `uqda help COMMAND` and `uqdactl help COMMAND` work offline, even when the
  daemon is stopped or its socket is inaccessible. All 14 real admin operations
  remain available; empty session/path results do not mean a command is broken.
- Local Unix admin-socket permission failures suggest rerunning the command
  with `sudo` on macOS/Linux, instead of incorrectly implying daemon failure.
  Root and Windows receive appropriate alternative guidance. No automatic
  elevation, password prompt or relaxed socket permissions is introduced.
- Permission failures return immediately without waiting for daemon startup.
  Status JSON remains valid and permission hints never echo invocation arguments
  that might contain peer passwords.
- The two-node CLI test harness waits for both independent daemon startups.

## Quick use

```sh
uqda help
uqda help addPeer
uqda help removePeer
uqda help test
sudo uqda status
sudo uqda info
sudo uqda peers
sudo uqda commands
```

Use `sudo` only when your macOS/Linux account cannot access the protected local
admin socket. Keep any custom endpoint and arguments when rerunning a command.
Do not make the socket world-accessible: the admin API has no authentication.

## Install and update

macOS:

```sh
brew update
brew install --cask Uqda/tap/uqda
# For an existing installation:
brew upgrade --cask Uqda/tap/uqda
```

Linux:

```sh
curl -fsSLo /tmp/uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh
sudo sh /tmp/uqda-install.sh install
# For an existing installation:
sudo sh /tmp/uqda-install.sh update
```

Verify release downloads against `SHA256SUMS`. Normal update, uninstall and
reinstall retain the configuration and private identity; explicit purge/zap
deletes that identity. Back it up privately before any purge. Installers remain
unsigned. Source builds require the complete checkout and local Ironwood patch.

## Compatibility and validation boundary

There is no new wire-format or private-group authentication change relative to
26.0.2. Members migrating from 26.0.1 or earlier still need the coordinated
Argon2id private-group upgrade described in the 26.0.2 release notes. Keep a
strong random shared password and the existing node identity.

Release validation covers core CI, all advertised CLI operations, actual denied
Unix socket tests on Linux/macOS, race regressions, native installer lifecycle
tests and mobile consumer builds. Refer to the release's linked Actions runs for
the executed architectures; cross-compilation does not certify native installation
or TUN reachability. Windows ARM64 packaging is build-only. A successful suite
does not prove the absence of every bug or constitute an independent security audit.
