# Installing on Ubiquiti EdgeOS

> **Requires EdgeOS 2.x.** This guide applies to Uqda integrated via the **vyatta-uqda** package bundled in [Uqda/Core](https://github.com/Uqda/Core) (`contrib/vyatta-uqda/`).

Uqda is supported on the Ubiquiti EdgeRouter using the **vyatta-uqda** package.

Perform installation steps over SSH by connecting to the EdgeRouter as the `ubnt` user, for example `ssh ubnt@192.168.1.1`, or another admin-level user if configured.

## Notes

Although your Uqda configuration will persist, the **vyatta-uqda** package itself does not survive an upgrade of the EdgeRouter firmware. You must re-add the repository GPG key (if you use an apt repository) and re-install the **vyatta-uqda** package after a system upgrade.

After upgrading firmware and reinstalling Uqda, use `load` to reload your configuration and then `commit` to make it effective again. Do not run `save` until after you have reloaded your configuration.

## Install the package

Download and copy the package onto the router. Once done, log into the router via SSH and use `dpkg` to install it:

```sh
sudo dpkg -i vyatta-uqda-x.x.xxx-mipsel.deb
```

## Generate configuration

Configuration for Uqda is generated automatically when you create an interface, for example as `tun0`:

```
configure
set interfaces uqda tun0
commit
save
```

At this point, Uqda will start running using default configuration, which includes automatic peer discovery of other Uqda nodes on the same network using multicast.

## Configuration

Once you have generated a configuration file as above, make configuration changes (like adding peers) by editing **`/config/uqda.tun0.conf`** (replace `tun0` with your interface name).

For example, if using `tun0`:

```sh
vi /config/uqda.tun0.conf
```

To make configuration changes effective, restart Uqda:

```
restart uqda tun0
```

## Masquerade

If you want to allow other IPv6 hosts on your network to communicate through Uqda, you can configure an IPv6 masquerade rule. All traffic sent from other hosts on the network through the Uqda interface will be NAT’d.

For example:

```
configure
set interfaces uqda tun0 masquerade from xxxx:xxxx:xxxx::/48
commit
save
```

If you have multiple IPv6 subnets, they can be configured individually by setting multiple `masquerade from` source ranges. Both private/ULA and public IPv6 subnets are acceptable.

## Default firewall config (example)

Use this as an example firewall configuration, which allows outgoing connections but prevents unexpected incoming ones, with the exception of ICMPv6 which will be allowed (for example with `tun0`):

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

## See also

- [Uninstall completely](uninstall.md)
- [Vyatta package README](../contrib/vyatta-uqda/README.md) (in the repository) for CLI reference.
