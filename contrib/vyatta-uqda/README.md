# Uqda for Ubiquiti EdgeOS / VyOS

This tree is **vendored inside** the [Uqda/Core](https://github.com/Uqda/Core) monorepo at `contrib/vyatta-uqda/`. Router `.deb` packages are built from the same checkout (see `.github/workflows/pkg.yml`, job `build-packages-router`).

### Introduction

This package provides uqda support on supported Ubiquiti EdgeOS 2.x, VyOS 1.3 and potentially other Vyatta-based routers.  It is integrated with the command line interface (CLI) allowing uqda to be configured through the standard configuration system.

### Compatibility

|                                  | Architecture | Tested |                      Notes                                    |
|----------------------------------|:------------:|:------:|:-------------------------------------------------------------:|
|    EdgeRouter X (ER-X/ER-X-SFP)  |    mipsel    |  Yes   | Tested with EdgeOS 2.0.9, requires EdgeOS 2.x                 |
|    EdgeRouter (ER-Lite/ER-4 etc) |    mips      |  No    | Requires EdgeOS 2.x                                           |
|    VyOS                          |    amd64     |  Yes   | Tested with VyOS 1.3-rolling-202101                           |
|    VyOS                          |    i386      |  No    |                                                               |

### Install / Upgrade

Either download or build a release and copy it to the router, then install/upgrade it:
```
sudo dpkg -i uqda-edgeos2x-x.x.x-xxxxxx.deb # EdgeOS
sudo dpkg -i uqda-vyos13-x.x.x-xxxxxx.deb   # VyOS
```

### Initial

Start by creating the default configuration on the interface (replacing `tunX` with your chosen TUN adapter):
```
configure
set interfaces uqda tunX
set interfaces uqda tunX description uqda
commit
```
This automatically generates a new private key and then populates the IPv6 address, public key and private key into the config.

### Configuration

Configuration changes should be made to `/config/uqda.tunX.conf` by hand. To make effective, restart uqda (replacing `tunX` with your chosen TUN adapter):
```
restart uqda tunX
```

### Masquerade

If you want to allow other IPv6 hosts on your network to communicate through uqda, you can configure an IPv6 masquerade rule. All traffic sent from other hosts on the network through the uqda interface will be NAT'd.

For example:
```
configure
set interfaces uqda tun0 masquerade from xxxx:xxxx:xxxx::/48
commit
```
If you have multiple IPv6 subnets, then they can be configured individually by setting multiple `masquerade from` source ranges. Both private/ULA and public IPv6 subnets are acceptable.

### CLI reference

Summarised EdgeOS/Vyatta-style commands (replace `tun0` with your interface).

**Create interface and persist**

```
configure
set interfaces uqda tun0
commit
save
```

Per-interface config file (example): **`/config/uqda.tun0.conf`**. Restart after edits:

```
restart uqda tun0
```

**Masquerade (NAT IPv6 LAN traffic via Uqda)**

```
configure
set interfaces uqda tun0 masquerade from xxxx:xxxx:xxxx::/48
commit
save
```

**Firewall (IPv6 inbound / local — example names `UQDA_IN`, `UQDA_LOCAL`)**

```
configure

set firewall ipv6-name UQDA_IN default-action drop
set firewall ipv6-name UQDA_LOCAL default-action drop

set firewall ipv6-name UQDA_IN rule 10 action accept
set firewall ipv6-name UQDA_IN rule 10 state established enable
set firewall ipv6-name UQDA_IN rule 10 state related enable

set firewall ipv6-name UQDA_IN rule 20 action drop
set firewall ipv6-name UQDA_IN rule 20 state invalid enable

set firewall ipv6-name UQDA_IN rule 30 action accept
set firewall ipv6-name UQDA_IN rule 30 protocol icmpv6

set firewall ipv6-name UQDA_LOCAL rule 10 action accept
set firewall ipv6-name UQDA_LOCAL rule 10 state established enable
set firewall ipv6-name UQDA_LOCAL rule 10 state related enable

set firewall ipv6-name UQDA_LOCAL rule 20 action drop
set firewall ipv6-name UQDA_LOCAL rule 20 state invalid enable

set firewall ipv6-name UQDA_LOCAL rule 30 action accept
set firewall ipv6-name UQDA_LOCAL rule 30 protocol icmpv6

set interfaces uqda tun0 firewall in ipv6-name UQDA_IN
set interfaces uqda tun0 firewall local ipv6-name UQDA_LOCAL

commit
save
```

See also **[docs/install-edgeos.md](../../docs/install-edgeos.md)** in the Core repository.

