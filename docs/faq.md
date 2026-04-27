# Uqda FAQ

## My private key or full config was exposed — what do I do?

Generate a **new** keypair and configuration; your Uqda IPv6 address will change. Step-by-step commands and merge checklist: **[Key rotation after a leak](key-rotation.md)**.

## Is Uqda anonymous?

No. Uqda provides end-to-end encryption between nodes, but it does **not** provide anonymity. Your IP address and public key are visible to peers you connect to directly. For anonymity, use Tor or a VPN in addition to Uqda.

## Has Uqda been audited for security?

Uqda has not undergone a formal third-party security audit. The project is currently alpha-stage. Do not rely on it for critical or high-security applications.

## Does Uqda hide my traffic from my ISP?

Uqda encrypts traffic between nodes end-to-end, but your ISP can still see that you are sending encrypted packets to your peers. The content is hidden, but the fact of communication is not.

## Should I run a firewall?

Yes. Uqda does not replace a firewall. All Uqda nodes on the network can attempt to reach your node's Uqda IPv6 address. Configure your OS firewall (ip6tables, pf, Windows Firewall) to allow only the services you intend to expose.

For **HTTP/HTTPS and static sites** on the overlay, see **[Hosting a website on Uqda](hosting-on-uqda.md)**.

## Is the admin socket secure?

The admin socket (used by **uqdactl**) has no built-in authentication. By default it uses a Unix socket at **`/var/run/uqda.sock`**, which is protected by filesystem permissions. If you change **AdminListen** to a TCP address, restrict access with your firewall — never expose the admin socket to untrusted networks.

## What is the address range?

Uqda uses the **200::/7** IPv6 range. Addresses are derived from your ed25519 public key and are permanent and location-independent.

## Can I peer over IPv4?

Yes. Uqda peering connections (TLS, QUIC, WebSocket) can run over both IPv4 and IPv6 networks, even though the overlay is IPv6.

## What is the minimum MTU?

**1280** bytes. The default **IfMTU** is the largest supported value for your platform (often up to **65535**; OpenBSD uses a lower platform maximum).

## Is Uqda compatible with Yggdrasil nodes?

Uqda is a fork/rebrand of Yggdrasil-go. Protocol compatibility with Yggdrasil public network nodes depends on the wire protocol version. Nodes running the same wire protocol version (e.g. 0.5) can peer.

## Where do I find public peers?

See the community peer lists or add your own peers using the **Peers** option in your **uqda.conf** configuration file.
