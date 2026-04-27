# Rotating your keys after a leak or compromise

If **anyone may have seen your `PrivateKey`**, your **PEM key file**, or a **full copy of `uqda.conf`**, treat the identity as compromised. Uqda has no “remote revoke” — you recover by **generating a new keypair** and **replacing the configuration** your node uses. Your **Uqda IPv6 address and `/64` subnet** will change, because they are derived from the public key.

**This is intentional:** a new key is a clean break from the leaked material.

---

## What changes when you rotate

| Item | After rotation |
|------|----------------|
| **PrivateKey** / PEM file | New random material |
| **Public key** (visible to peers) | New value; use `uqda -useconffile … -publickey` to print it |
| **Your Uqda IPv6 address** | New (`uqda … -address` or `uqdactl getSelf` once running) |
| **Your routed `/64` prefix** | New (first byte `200`→`300` pattern; see [Configuration](configuration.md#advertising-a-prefix)) |
| **Peers / Listen / multicast** | You copy these from the old file if you want the same topology |
| **Other people’s keys in `AllowedPublicKeys`** | Usually unchanged (those are *their* public keys) |

Anyone who **restricted inbound peerings to your old public key** must **update** their config to your **new** public key (or relax the list). You do **not** need to change your list of their keys unless *their* keys were also compromised.

---

## Recommended procedure (static config file)

**1. Stop Uqda** (adjust for your init system):

```sh
sudo systemctl stop uqda          # Linux systemd
# or: sudo launchctl unload /Library/LaunchDaemons/uqda.plist   # macOS launchd
```

**2. Keep a sealed backup of the compromised file** (for forensics only — do not run it on the network again):

```sh
sudo cp /etc/uqda.conf /root/uqda.conf.compromised-$(date +%Y%m%d).bak
sudo chmod 600 /root/uqda.conf.compromised-*.bak
```

Use your real path if the file is **`/etc/uqda/uqda.conf`**, **`%ALLUSERSPROFILE%\Uqda\uqda.conf`**, or a path in your home directory.

**3. Record the old identity** (optional, for telling peers or revoking references):

```sh
sudo uqda -useconffile /root/uqda.conf.compromised-YYYYMMDD.bak -publickey
sudo uqda -useconffile /root/uqda.conf.compromised-YYYYMMDD.bak -address
```

**4. Generate a brand-new baseline config:**

```sh
sudo uqda -genconf > /tmp/uqda.conf.new
sudo chmod 600 /tmp/uqda.conf.new
```

**5. Merge non-secret settings** from the backup into **`/tmp/uqda.conf.new`**:

Copy over (as applicable): **`Peers`**, **`InterfacePeers`**, **`Listen`**, **`MulticastInterfaces`**, **`NodeInfo`**, **`IfName`**, **`IfMTU`**, **`AllowedPublicKeys`** (others’ keys), **`AdminListen`**, and any other tuning you rely on.

Do **not** copy **`PrivateKey`** or **`PrivateKeyPath`** from the old file.

**6. If you used `PrivateKeyPath`:** delete or archive the old PEM file, export a new key, and point **`PrivateKeyPath`** at the new file only:

```sh
sudo uqda -useconffile /tmp/uqda.conf.new -exportkey > /path/to/new-uqda.key
sudo chmod 600 /path/to/new-uqda.key
```

Then edit the active config so it uses **`PrivateKeyPath`** and does **not** embed the hex **`PrivateKey`** in the same file (see [Configuration](configuration.md#exporting-keys-to-external-files)).

**7. Install the new config and start Uqda:**

```sh
sudo cp /tmp/uqda.conf.new /etc/uqda.conf    # or your distro path
sudo chmod 600 /etc/uqda.conf
sudo systemctl start uqda
```

**8. Update everything that assumed the old address**

Examples: **firewall rules**, **split routing**, **DNS or bookmarks** to your old Uqda IP, **radvd** prefix on a LAN, **documentation** you gave to friends, and **peer lists** on other nodes that pinned your old public key.

**9. Confirm the new identity:**

```sh
uqdactl getSelf
uqda -useconffile /etc/uqda.conf -publickey
```

---

## One-liners useful during rotation

```sh
# New config template (stdout only — redirect to a file to save)
uqda -genconf

# Show address / subnet / public key from a file (no TUN required)
uqda -useconffile /etc/uqda.conf -address
uqda -useconffile /etc/uqda.conf -subnet
uqda -useconffile /etc/uqda.conf -publickey

# Pretty-print / normalise after edits
uqda -normaliseconf -useconffile /etc/uqda.conf
```

---

## `-autoconf` is not a fix for compromise

**`-autoconf`** generates **new random keys on every start**. That is fine for quick LAN tests, not for replacing a stable node identity: your address would change every reboot and peer pinning would be impossible. For recovery after a leak, use a **new static config** as above.

---

## See also

- [Configuration](configuration.md) — generating config, exporting keys, normalising.
- [Configuration reference](configuration-reference.md) — **`PrivateKey`**, **`PrivateKeyPath`**, **`AllowedPublicKeys`**.
- [FAQ](faq.md) — admin socket exposure and anonymity notes.
