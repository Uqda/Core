# Configuration

Uqda accepts HJSON or JSON. Use `uqda -genconf` for a new identity and defaults;
add `-json` for JSON. Save the result once as `uqda.conf` with restrictive
permissions. Never redirect generated output over an existing identity.

`uqda -useconffile uqda.conf -checkconf` validates without starting the node.
`-normaliseconf` prints normalized configuration; write it to a temporary file,
validate it and back up the original before replacement.

`PrivateKey` or `PrivateKeyPath` provides persistent identity. Missing identity
in a configuration loaded by the daemon is an error; `-autoconf` explicitly selects an ephemeral
identity. External key paths must remain accessible to the service account.

`Peers` specifies outbound peer URIs; `Listen` specifies inbound listeners.
`MulticastInterfaces` controls local discovery. `AdminListen` controls the local
administration socket; it has no authentication. `IfName: none` disables TUN.
`GroupPassword` restricts encrypted sessions to peers with the same secret,
while relay/transport peering remains separate.

Starting in 26.0.2, private-group authentication uses Argon2id. Upgrade **every
member of a private group together** from 26.0.1 or earlier. Mixed old/new members
may still show a transport peer, but cannot exchange encrypted session traffic.
The password and identity need not change. Empty/public mode remains compatible
with upstream. There is no legacy-password authentication fallback.

Use a strong randomly generated group secret, shared only with trusted members.
Derivation uses 64 MiB transient memory, three passes and four lanes once per node
startup, not per incoming packet. Low-memory devices must budget that startup
memory. The deterministic protocol salt lets group members derive the same key;
it cannot prevent precomputation against weak passwords reused between groups.

The upstream [Yggdrasil configuration reference](https://yggdrasil-network.github.io/configurationref.html)
explains shared network options. Uqda binary names, paths and validation behavior
are documented here; upstream installation instructions are not Uqda instructions.
