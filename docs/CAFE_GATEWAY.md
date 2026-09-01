# UQDA Café Gateway

The café profile lets an operator run one UQDA node and offer a dedicated
Wi-Fi network to untrusted visitors. Visitors do not install UQDA: their normal
IPv6 stack receives an address from the gateway's routed UQDA `/64` and an
UQDA route through Router Advertisements.

This profile extends the [home gateway](HOME_GATEWAY.md). A public hotspot has
a different threat model: every client, including a malicious one, must be
treated as untrusted.

## Security boundary

The café profile applies four controls together:

1. Wi-Fi AP isolation blocks direct client-to-client traffic at layer 2.
2. The hotspot has its own IPv4 subnet and UQDA IPv6 prefix.
3. The gateway firewall accepts only DHCP, DNS, and essential IPv6 control
   traffic from visitors; SSH, web administration, and the host itself are
   blocked.
4. Forwarding to private/management address ranges on the upstream side is
   blocked while public Internet and UQDA destinations remain routable.

AP isolation is necessary but not sufficient. Some drivers do not implement
it correctly, and it does not replace the layer-3 firewall. Management should
use a separate wired interface or VLAN that is never assigned to the café
network.

## Recommended topology

Use a dedicated Raspberry Pi 5, small Linux appliance, or supported OpenWrt
router. Connect its WAN Ethernet port to the café's Internet router and use a
separate Wi-Fi radio for visitors. Do not bridge the visitor SSID into the
point-of-sale, cameras, printers, staff Wi-Fi, or router-management network.

For a commercial deployment, prefer:

- a dedicated AP-capable radio with stable Linux/OpenWrt drivers;
- at least 4 GB RAM for a busy Raspberry Pi deployment;
- a wired management interface or tagged management VLAN;
- an uninterruptible power supply; and
- a spare, preconfigured replacement device.

## Install on Raspberry Pi OS, Debian, or Ubuntu

Install UQDA first, then the gateway dependencies:

```sh
sudo apt update
sudo apt install network-manager radvd iw nftables
curl -fsSLo uqda-gateway \
  https://raw.githubusercontent.com/Uqda/Core/main/contrib/gateway/uqda-gateway
chmod +x uqda-gateway
sudo install -m 0755 uqda-gateway /usr/local/sbin/uqda-gateway
```

Identify interfaces while connected through Ethernet or a local console:

```sh
ip -brief link
nmcli device status
nmcli -f WIFI-PROPERTIES.AP device show wlan0
```

Create a strong visitor password without adding it to shell history:

```sh
umask 077
read -r -s -p "Café Wi-Fi password: " UQDA_WIFI_PASSWORD; echo
printf '%s\n' "$UQDA_WIFI_PASSWORD" > /tmp/uqda-cafe-password
unset UQDA_WIFI_PASSWORD
```

Review the read-only plan, then explicitly acknowledge public operation:

```sh
uqda-gateway plan --profile cafe --wan eth0 --lan wlan0 \
  --ssid Cafe-UQDA --country DE

sudo uqda-gateway apply --profile cafe --public \
  --wan eth0 --lan wlan0 --ssid Cafe-UQDA --country DE \
  --password-file /tmp/uqda-cafe-password
rm -f /tmp/uqda-cafe-password
```

The country code must match the physical location. The command refuses to
guess interfaces and refuses café mode without `--public`.

## Install on OpenWrt

The OpenWrt profile is experimental until validated on the exact router model.
It uses a separate firewall zone with rejected input/forwarding, explicit
DHCP/DNS/ICMPv6 allowances, WAN forwarding, and wireless client isolation.

```sh
uqda-gateway plan --backend openwrt --profile cafe \
  --wan wan --lan radio0 --ssid Cafe-UQDA --country DE

uqda-gateway apply --backend openwrt --profile cafe --public \
  --wan wan --lan radio0 --ssid Cafe-UQDA --country DE \
  --password-file /root/uqda-cafe-password
rm -f /root/uqda-cafe-password
```

Keep a wired recovery connection. Confirm that `wan` is the Internet firewall
zone and `radio0` is the intended visitor radio before applying.

## Acceptance tests before opening

Use two ordinary visitor devices and one management device:

- both visitors receive Internet access and an address from the UQDA `3xx:`
  subnet;
- both can reach a known UQDA peer;
- neither visitor can ping, discover, or open a service on the other;
- neither can reach the gateway's SSH/web administration or upstream private
  networks;
- the separate management device can still administer the gateway;
- configuration survives a reboot; and
- `sudo uqda-gateway rollback` restores the previous configuration.

Repeat these tests after every firmware, kernel, Wi-Fi driver, NetworkManager,
OpenWrt, or UQDA upgrade.

## Operations and privacy

- Rotate the visitor password regularly and after suspected misuse.
- Publish an acceptable-use and privacy notice appropriate to local law.
- Keep only the minimum logs necessary for reliability and legal obligations;
  never claim UQDA makes visitors anonymous.
- Set bandwidth/client limits on the AP or upstream router for a busy venue.
- Monitor aggregate health, disk space, peer state, and abuse reports without
  inspecting visitor content.
- Do not expose the UQDA administrator socket over TCP.

A captive portal is optional and is deliberately outside this script. If the
venue needs terms acceptance, vouchers, or per-user limits, integrate a mature
portal such as an OpenWrt-supported solution on the isolated visitor zone and
test that it does not bypass IPv6/UQDA firewall policy.

## Emergency shutdown

To disable the visitor network and restore the configuration saved before the
most recent apply:

```sh
sudo uqda-gateway rollback
```

For an immediate NetworkManager shutdown without changing saved state:

```sh
sudo nmcli connection down uqda-gateway
```
