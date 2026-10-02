# Terminal command center

Release 26.0.4 adds `uqda find` for read-only public peer discovery.
See [public peer suggestions](public-peers.md) for usage and security boundaries.
This command requires 26.0.4; it is not in the 26.0.3 binaries.

Release 26.0.2 adds one everyday entry point: `uqda`. Older installations can
still use `uqdactl doctor`; the shortcuts require 26.0.2 or newer.

```text
uqda                  Node health and connection status
uqda peers            Connected peers and traffic
uqda test 200:1234::1  Test another node's Uqda IPv6 address
uqda info             Node address and public identity
uqda version          Installed release
uqda help             Complete command guide
```

Release 26.0.3 includes a complete help guide for all built-in admin operations, daemon
configuration options, examples, runtime-versus-persistent peer changes, security
notes and exit codes. `uqda help addPeer`, `uqda help removePeer`, `uqda help test`
and `uqdactl help getNodeInfo` show individual command details without contacting
the daemon. Help topics `install`, `update` and `uninstall` explain the external
platform tools; they do not add fake `uqda install` or `uqda uninstall` operations.
These expanded help pages require 26.0.3; they are not in the 26.0.2 binaries.

The 26.0.3 controller also distinguishes local admin-socket permission
denials from an unavailable daemon and suggests rerunning the same command with
`sudo` on macOS/Linux (for example `sudo uqda info`). Keep your original custom
endpoint and arguments. It does not elevate automatically, change socket
permissions, or print invocation arguments that may contain secrets. A root
permission denial instead suggests checking ownership and OS security policy.
This guidance is not included in the 26.0.2 binaries. On 26.0.2,
use `sudo uqda`, `sudo uqda info` and `sudo uqda peers` when your account cannot
access the protected local admin socket. Do not make that socket world-accessible:
the admin API has no authentication.

`uqda` without arguments checks the existing daemon; it does not start a new node,
modify configuration, or generate an identity. Keep `uqda` and `uqdactl` together
in the installed binary directory. The launcher never searches PATH for an
unrelated controller. On Windows, use `uqda.exe` from CMD or `./uqda.exe` from
PowerShell when running a local build.

Output uses aligned fields, peer cards and ASCII status labels, without requiring
color, Unicode fonts or terminal escape sequences. Peer URI credentials and raw
peer error messages are omitted from the human-readable view. JSON remains the
admin API's raw response and may contain private peer information: treat it as
sensitive.

```sh
uqda status --json
uqda peers --endpoint tcp://127.0.0.1:9001
```

Options may appear before or after the command in `uqdactl`. With `uqda`, put the
command first: leading flags retain their existing daemon meaning.

## Advanced administration

`uqda commands` lists the running node's actual supported admin operations.
Use `uqdactl <operation> name=value` for advanced operations. The controller
rejects unknown operations, unsupported parameters, malformed values and duplicate
parameter names rather than silently ignoring them. Existing working diagnostic
commands are retained for compatibility; they are not shown in the daily menu.

`uqda help advanced` shows daemon flags. Starting a node still requires explicit
configuration, for example `uqda -useconffile uqda.conf`. Service launch flags are
unchanged. Identity/configuration inspection flags require a configuration file
or stdin configuration; missing input is an error, not a successful no-op.

Exit status is 0 for success, 1 for an operational failure and 2 for invalid usage.
Health checks may return 0 with warnings; read their status or JSON report.

## Tests

`tests/cli/commands.py` builds no virtual network interface and does not touch
installed configuration. It connects two real local private-group daemons and
tests every advertised admin operation, short commands, JSON and error paths.
GitHub Actions runs it on Windows, Linux and macOS. Existing network/data-plane
and packaging tests remain separate gates; this test does not claim to prove
TUN connectivity or complete OS installation.
