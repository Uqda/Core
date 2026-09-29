# Uqda doctor

Run a quick, read-only health check with either easy-to-remember command:

```sh
uqdactl doctor
uqdactl status
```

In Windows PowerShell, use `uqdactl.exe doctor` (or
`.\uqdactl.exe doctor` when running it from its own directory). On macOS
and Linux, the same `uqdactl doctor` spelling works when installed in `PATH`.

For scripts, use `uqdactl -json doctor`. For a non-default administration
socket, put the option before the command, for example:

```sh
uqdactl -endpoint=unix:///run/uqda/admin.sock doctor
```

The report checks whether the admin API responds, whether the node reports a
valid Ed25519 public key and IPv6 address, how many peers are connected, and
whether the TUN interface is enabled. `PASS` means the observed check succeeded;
`WARN` flags a condition that may be intentional (for example, router-only mode);
`FAIL` means a required check could not be completed. The command exits with
status 1 for a failure, otherwise 0. JSON uses the same `pass`, `warn`, and
`fail` statuses.
When the daemon has just started, the first admin check retries for up to five
seconds while its socket becomes ready.

The report deliberately omits configuration contents, peer URIs, peer error
messages, keys, and detailed endpoint addresses. It does not change service
state or configuration. The API has no authentication, so do not expose it
to an untrusted network. A responding admin API and an up peer do not prove
end-to-end IPv6 reachability, identity persistence after restart, or firewall
safety; use the two-node integration checks for those claims.
