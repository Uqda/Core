# Configuration

Uqda accepts HJSON or JSON. Use `uqda -genconf` for a new identity and defaults;
add `-json` for JSON. Save the result once as `uqda.conf` with restrictive
permissions. Never redirect generated output over an existing identity.

`uqda -useconffile uqda.conf -checkconf` validates without starting the node.
`-normaliseconf` prints normalized configuration; write it to a temporary file,
validate it and back up the original before replacement.

`PrivateKey` or `PrivateKeyPath` provides persistent identity. Missing identity
in a loaded configuration is an error; `-autoconf` explicitly selects an ephemeral
identity. External key paths must remain accessible to the service account.

`Peers` specifies outbound peer URIs; `Listen` specifies inbound listeners.
`MulticastInterfaces` controls local discovery. `AdminListen` controls the local
administration socket; it has no authentication. `IfName: none` disables TUN.
`GroupPassword` restricts encrypted sessions to peers with the same secret,
while relay/transport peering remains separate.

The upstream [Yggdrasil configuration reference](https://yggdrasil-network.github.io/configurationref.html)
explains shared network options. Uqda binary names, paths and validation behavior
are documented here; upstream installation instructions are not Uqda instructions.
