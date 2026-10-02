# Terminal command center

The next source build adds one everyday entry point: `uqda`. Installed release
26.0.1 still uses `uqdactl doctor`; these shortcuts require the new build.

```text
uqda                  Node health and connection status
uqda peers            Connected peers and traffic
uqda test 200:1234::1  Test another node's Uqda IPv6 address
uqda info             Node address and public identity
uqda version          Installed release
uqda help             Short command guide
```

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
