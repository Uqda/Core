# Uqda on Umbrel

The `uqda-network` community-store package runs the verified Core 26.0.4
release with an Arabic/English browser dashboard. The root
`umbrel-app-store.yml` makes this repository usable as a community app store
once the package image is published and its digest is committed.

## Install

1. In Umbrel's App Store, open Community App Stores and add
   `https://github.com/Uqda/umbrel-app-store`. Menu wording varies by umbrelOS version.
2. Install **Uqda Network** and open it.
3. Sign in using the app password shown by Umbrel in the app's details.
4. Enter a trusted peer URI, for example `tls://peer.example:443`, and the
   strong group password used by your other private-group members.
5. Save and reconnect. Check the node address, transport peers and TUN state.

New nodes start with a unique persistent identity, no peers, no incoming
listeners, multicast disabled, and a random private-group secret. They remain
isolated until configured. Enter your shared group secret to join an existing
group; a blank password keeps the current secret. Passwords and private keys
are never returned by the dashboard API.

Public mode requires a checked confirmation and clears `GroupPassword`.
That permits public encrypted sessions and may expose services listening on
the host's Uqda IPv6 address. A connected transport peer alone is not evidence
of a working private-group session. Verify connectivity from another trusted
node using the existing Uqda network diagnostics.

## What this app is useful for

Use the dashboard to join your Umbrel host to your trusted Uqda devices with a
persistent encrypted IPv6 identity, manage private-group settings, and inspect
the local node without editing configuration files. It is not a reverse proxy,
an anonymity service, or a switch that publishes every Umbrel app.

After configuration, run `sudo uqda test UMBREL_UQDA_IPV6` from another trusted
group member (replace the placeholder with the address shown in the dashboard).
Then test the actual application's IPv6 endpoint separately. The interface and
peer health checks only describe local/transport readiness; they do not establish
remote application availability. The dashboard's bilingual setup guide explains
these steps and the backup/removal implications before changing a service.

The dashboard supports TLS/TCP/QUIC/WS/WSS peer URLs without embedded
credentials, and optional TLS/TCP/QUIC incoming listeners on ports >=1024.
It never adds public peers automatically. Advanced credential-bearing or
unsupported transport settings are preserved and make browser editing
unavailable. Incoming listener configuration does not change router or host
firewall settings.

## Your services over Uqda

For the broader dashboard/app-gateway workflow, research evidence and remaining
mobile/TLS/native-client acceptance work, see [Umbrel remote access](umbrel-remote-access.md).

This workflow is a source enhancement, not part of the already published
`26.0.4-umbrel.2` image. Distribution requires a new immutable wrapper version
and digest-pinned community-store update after validation; the existing image
and installed devices are not modified by this change.

The dashboard includes an explicit service-access workflow, in Arabic and
English: connect another device, prepare one installed service, and verify it
from that other device. It is a **service address book**, not an app installer
or publishing switch. Select HTTPS, HTTP, SSH or generic TCP, give the service
a name and enter its real listening port. Up to twelve entries are saved.

For example, if an existing web service listens on this host's Uqda IPv6 at
port 8443 with HTTPS enabled, save `My files / HTTPS / 8443`. The dashboard
generates `https://[YOUR_UQDA_IPV6]:8443/`, a `sudo uqda test` network command,
and a bounded `curl --head` command. No service-specific ports are guessed.
SSH instructions use an existing account (`USER` is a placeholder), and TCP
instructions use `nc`. Install these diagnostic clients separately if absent.
HTTPS diagnostics retain certificate verification: a raw IPv6 URL may not
match your certificate. Configure a suitable certificate/name rather than
disabling verification. A failed HEAD request can also mean the app does not
support HEAD; test its intended authenticated use separately.

Saving entries does **not** start services, expose IPv4-only Docker apps,
install a reverse proxy, open firewall/router ports, enable forwarding, add
users or modify Core identity/settings. Removing an entry only removes its
instructions; it does not uninstall, stop or secure the underlying service.
Configure that service's IPv6 binding, authentication and narrowly scoped
firewall rules separately. Do not blindly bind all applications to `::`.

**Check local TCP** makes one bounded TCP connection to a saved port on this
daemon's own literal Uqda IPv6 identity. It sends no application payload and
cannot accept an arbitrary host, port, URL, shell command or credential.
Success means only that TCP accepted a connection locally. The UI always
keeps **Remote access: not verified**: local listening, an UP transport peer,
application health, TLS/authentication and end-to-end reachability are distinct.
Use the generated commands on another compatible Uqda-connected device, then
verify the actual app login and operation. Opening this dashboard on an iPad
does not itself connect the iPad to the overlay.

Prefer a private group for your own devices. Each member needs the same strong
group secret and a reachable transport peer; a public-mode node cannot simply
join a private-group session. Public mode keeps overlay encryption but does
not provide your group's access restriction. The UI warns about IPv6 services
potentially reachable by other public nodes if permitted by the firewall.

Web addresses are displayed as copyable text, not auto-opened links. Cookies
can cross ports on the same HTTP host, so use a separate browser profile for
an untrusted co-hosted app and HTTPS where available. No password or secret is
included in generated addresses or commands. All service mutation/probe routes
require the existing authenticated session, same Origin and session proof.

Entries live in `data/config/services.json`, root-owned mode 0600. Include this
file in app-data backups. They survive Core restarts and container recreation,
but deleting app data removes the address book as well as the node identity.
Revision checks reject stale edits. Probe results are deliberately ephemeral
and cleared when the identity, network configuration or service list changes.

## Architecture and permissions

`core` uses host networking, `/dev/net/tun` and `NET_ADMIN` because this
app connects the Umbrel host to the encrypted IPv6 overlay. It creates the
`uqda0` host interface for new configurations. It does not configure forwarding,
change firewall rules, publish other apps, or mount the Docker socket.

`dashboard` runs as UID/GID 1000 with all Linux capabilities dropped,
no-new-privileges and a read-only filesystem on the Docker bridge network.
It has no access to `/etc/uqda` or the raw admin socket. Only a dedicated
group-restricted Unix control socket is shared with it. The supervisor exposes
status, validated settings, restart and bounded service-address-book operations, not arbitrary admin requests
or shell commands. The raw Core admin socket stays in the core container's
private runtime directory and is never exposed over TCP.

Umbrel's authenticated app proxy routes host port 8926 to dashboard port 8080.
No raw dashboard port is published by the package. The dashboard also requires
the stable per-install `APP_PASSWORD`, with constant-time comparison, bounded
login attempts, eight-hour sessions, HttpOnly/SameSite cookies, Origin and CSRF
checks on mutations, a restrictive CSP, and no third-party scripts or fonts.
Authenticated reads also require a random session proof held in origin-scoped
browser `sessionStorage`, not another cookie. Cookies ignore port numbers, so a
co-hosted app receiving the cookie cannot bootstrap this proof or read status.
Reloading the same tab preserves access; a new tab or blocked session storage
may require signing in again. Logout clears the tab's proof.
Use HTTPS when available; plain HTTP provides no transport protection for
browser passwords. Stopping/recreating the dashboard invalidates its sessions.

Overlay reachability does not guarantee access to Umbrel apps: services must
listen on IPv6 and their firewall must permit the connection. IPv4-only
Docker port publishing may need separate administrator configuration. No
automatic port exposure or network-wide discovery is implemented.

## Persistence, updates and recovery

Persistent configuration and identity live in
`${APP_DATA_DIR}/data/config/uqda.conf`, owned by root with mode 0600 in a
0700 directory. The supervisor refuses concurrent access to the same config.
Only the dedicated runtime control socket lives in `data/control`, excluded
from backups. It is regenerated on startup. The config, previous config and
pre-Umbrel config backups are included in backups and contain secret material.

On first import, Core parses/normalizes an existing HJSON/JSON configuration;
its identity, group protection, allowlists and advanced settings are retained.
Only `AdminListen` is moved to the container-private Unix socket. The first
imported bytes are backed up as `uqda.conf.before-umbrel`. External
`PrivateKeyPath` files must be inside the persisted config directory and use
container-accessible paths; the dashboard does not import arbitrary host files.
Do not copy an identity from a node that is still active.

Saving requires the current config revision, validates a temporary protected
file with Core, saves `uqda.conf.previous`, stops Core, atomically replaces the
configuration and waits for Core/TUN readiness. Startup failure restores the
previous bytes and restarts Core. A process/container crash between saving and
readiness may require manually restoring `uqda.conf.previous`; startup never
silently replaces or deletes a stored identity. Existing invalid configurations
fail startup with a generic validation error and must be repaired locally.

Use normal Umbrel app updates to replace the images while retaining app data.
Keep a protected backup before updating/removing/restoring. Check Umbrel's
uninstall/data-removal prompt: removing the stored data destroys the identity.
Do not use the Linux/systemd quick installer to update this container package.

## Build and validation

```sh
UQDA_TEST_BINARY=/path/to/verified/uqda python3 -m unittest discover -s tests/umbrel -v
node --check contrib/umbrel/web/app.js
node tests/umbrel/ui_state.cjs
python3 tests/umbrel/package_check.py --require-digest
```

The native tests use the actual release daemon with TUN disabled, covering
identity retention, secret redaction, private/public mode, invalid edits,
stale revisions, rollback, supervisor locking, authentication, logout, session
expiry, login limits, Origin/CSRF checks and the local control allowlist.
Service tests cover persistent metadata, stale edits, removal without changing
Core config, strict port/type validation, literal-address command generation,
arbitrary-target rejection and local-only probe semantics.

For Linux Docker with `/dev/net/tun`:

```sh
export UQDA_TEST_PASSWORD='a-long-local-test-password'
python3 tests/umbrel/docker_smoke.py
```

This gate checks actual host TUN, UID separation, lack of config/raw admin
access from the dashboard, config mode 0600, login, settings, restart and
container recreation with identity retained. It creates and deletes only its
own test volumes. Do not run it alongside an existing `uqda0` interface.

It also creates a disposable second node in an independent Docker network
namespace. The gate transfers 1 MiB over overlay IPv6 TCP in both directions,
checks the received SHA-256 hashes, exchanges UDP in both directions, rejects traffic with a different private-group
password even when transport peers connect, and checks recovery after restoring
the correct password. This demonstrates useful application traffic, not just a
green transport indicator. All listeners and the second node belong to the test.
The service-access fixture additionally registers an IPv6 HTTP service, checks
its local TCP listener through the bounded control API, and verifies the actual
HTTP body fetched from the separate node. Browser tests exercise registration,
safe text rendering, generated addresses, local-only results, RTL and mobile
layout; the saved entries must survive restart and container recreation.

## Testing without a physical Umbrel device

Physical hardware is not required for app lifecycle validation. Umbrel's official
`scripts/umbrel-dev` can run the platform in a privileged Linux Docker container
with its own Docker daemon, persistent data volume and systemd. In a separate,
disposable checkout of `getumbrel/umbrel` at the desired release:

```sh
npm run dev start
npm run dev production-mode
```

Install the dedicated community store through that instance's actual app store,
not by substituting `compose.dev.yml`. Exercise signup, Umbrel's app proxy,
dashboard login, settings, private overlay traffic, app restart, update with data
retained, and uninstall/reinstall. Use `npm run dev restart` to test a development
instance restart. This is not a test of a production kernel/bootloader reboot.
The official script requires native Linux Docker networking (or an appropriate
macOS/WSL2 environment); a Windows Docker client alone is not sufficient.

For production boot/reboot behavior, install the official umbrelOS ISO in an
EFI virtual machine with at least 4 GB RAM and a 32 GB virtual disk. Install only
onto a new disposable virtual disk, never an existing server disk. Test the app
through the real Umbrel UI and restart the VM. Repeat native arm64 runtime tests
separately; amd64 VM success does not establish arm64 support.

References: [official development script](https://github.com/getumbrel/umbrel/blob/2.0.0/scripts/umbrel-dev)
and [official VM installation guide](https://umbrel.com/support/install-umbrelos-on-your-own-hardware/installing-umbrelos-in-a-virtual-machine).

The separate `umbrel-platform.yml` gate installs the digest-pinned published
package and a current-source image through the pinned official Umbrel 2.0 app
manager. The latter is built inside the disposable development platform, pushed
only to a loopback-bound temporary registry, then pinned by digest in the test
store. It is not published to GHCR or substituted in the public store.
The gate starts LAN ingress
on the normal internal server port (the upstream test factory's default port 0
intentionally disables it). It checks anonymous gateway rejection, actual owner
session cookies plus the separate dashboard password, restart, a manifest-only
update, and uninstall/fresh-install identity replacement. The source matrix also
switches to the previous published wrapper then upgrades to the current published
digest, checking identity retention and the new read-auth boundary. Its initial
install also exercises a freshly rebuilt source image. The published
matrix's manifest-only update is not a binary upgrade test.
Consult the run result before claiming these gates
passed; the test source alone is not evidence.

The CI fixture limits the initial empty development service's pre-test stop to
30 seconds instead of upstream's 15 minutes. Systemd may force-stop that fixture
after the deadline; this is not proof of graceful production OS shutdown. Actual
Uqda app restart/update operations use the unchanged official app manager.

Plain Docker tests do not establish actual Umbrel app-store/proxy behavior. A
development-instance test does not establish production boot or hardware-specific
behavior. Record these results separately; neither requires buying a device.

The Umbrel workflow runs both gates, builds amd64/arm64 wrappers, and publishes
new versioned wrapper tags to the existing public `ghcr.io/uqda/core` package on
authorized same-repository branch/main pushes. Existing tags are not overwritten;
increment `contrib/umbrel/VERSION`, manifest version and compose tags for wrapper
updates. Core release tags and binaries are independently versioned.

After publication, commit the multiarchitecture digest into **both** compose
image references. `export_store.py` produces a minimal digest-pinned store ZIP:

```sh
python3 contrib/umbrel/export_store.py --digest sha256:ACTUAL_INDEX_DIGEST --output uqda-umbrel-store.zip
```

Before claiming Umbrel support is verified, test through an actual Umbrel
installation: add the store, install, confirm both auth layers, save real state,
connect a trusted peer, verify encrypted overlay traffic, restart the device,
update with identity retained, and check both supported architectures. Native
daemon/Docker tests do not establish Umbrel lifecycle or arm64 runtime behavior.
