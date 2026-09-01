# UQDA Home Gateway

An UQDA home gateway runs one UQDA node on a router or small computer and
shares it with phones, tablets, televisions, and computers over Wi-Fi. Client
devices do not need the UQDA application. They receive an address from the
node's routed UQDA `/64` using standard IPv6 Router Advertisements and SLAAC.

## What it does

- ordinary websites continue to use the normal ISP connection;
- UQDA IPv6 destinations are routed through the encrypted overlay;
- the gateway supplies normal IPv4 Internet sharing to Wi-Fi clients; and
- UQDA's administrator API remains on the local Unix socket.

This is not an anonymity service and it is not an Internet exit VPN. Sending
all public Internet traffic through a remote machine requires a separately
operated and trusted exit gateway, DNS policy, abuse controls, and additional
firewall rules.

## Recommended hardware

| Profile | Suggested device | WAN | Home devices | Status |
|---|---|---|---|---|
| Raspberry Pi | Raspberry Pi 4 or 5, 2 GB+, quality power supply | Ethernet | onboard Wi-Fi | primary reference profile |
| Linux appliance | x86-64/ARM64 mini PC, 2 GB+, AP-capable Wi-Fi | Ethernet | Wi-Fi | supported through NetworkManager |
| OpenWrt router | supported CPU, TUN, at least 128 MB RAM and free storage | existing `wan` zone | one radio | experimental until tested per model |

A separate USB Wi-Fi adapter may be preferable when the device's onboard
radio cannot operate reliably as an access point. Never attempt initial setup
over the same Wi-Fi interface that will be converted into the access point;
use Ethernet or a local console.

## Raspberry Pi OS, Debian, or Ubuntu

Raspberry Pi OS uses NetworkManager. Debian or Ubuntu installations must also
have NetworkManager controlling the selected Wi-Fi device. Install UQDA first,
then install the gateway dependencies:

```sh
sudo apt update
sudo apt install network-manager radvd
curl -fsSLO https://github.com/Uqda/Core/releases/latest/download/uqda-gateway
chmod +x uqda-gateway
sudo install -m 0755 uqda-gateway /usr/local/sbin/uqda-gateway
```

List interfaces and Wi-Fi capabilities:

```sh
ip -brief link
nmcli device status
nmcli -f WIFI-PROPERTIES.AP device show wlan0
```

Create the password without putting it in shell history. Replace the interface
names and country code with the values for the device:

```sh
umask 077
read -r -s -p "Wi-Fi password: " UQDA_WIFI_PASSWORD; echo
printf '%s\n' "$UQDA_WIFI_PASSWORD" > /tmp/uqda-wifi-password
unset UQDA_WIFI_PASSWORD

uqda-gateway plan --wan eth0 --lan wlan0 --ssid Home-UQDA --country DE
sudo uqda-gateway apply --wan eth0 --lan wlan0 --ssid Home-UQDA \
  --country DE --password-file /tmp/uqda-wifi-password
rm -f /tmp/uqda-wifi-password
```

Use the correct two-letter regulatory country. The tool refuses to guess WAN
and Wi-Fi interfaces. It creates only a connection named `uqda-gateway` and
owned files under `/var/lib/uqda-gateway` so removal is targeted.

## OpenWrt

The OpenWrt profile requires working `uqda`, `uqdactl`, UCI, odhcpd, firewall,
TUN support, and sufficient device resources. UQDA's generic release installer
does not currently install OpenWrt packages automatically, so treat this
profile as experimental and validate it on the exact router model before home
use.

Copy `uqda-gateway` to `/usr/sbin`, identify the WAN firewall zone and radio,
then run:

```sh
ip link show
uci show wireless
uci show firewall | grep "\.name='wan'"

uqda-gateway plan --backend openwrt --wan wan --lan radio0 --ssid Home-UQDA
uqda-gateway apply --backend openwrt --wan wan --lan radio0 \
  --ssid Home-UQDA --country DE --password-file /root/uqda-wifi-password
rm -f /root/uqda-wifi-password
```

The profile creates an isolated `192.168.82.0/24` LAN, WPA2/WPA3 mixed-mode
Wi-Fi, DHCP, IPv6 RA/DHCPv6, and forwarding from that LAN to the selected WAN
zone. UQDA supplies the routed IPv6 `/64`.

## Check and recover

```sh
sudo uqda-gateway status
sudo uqdactl getSelf
sudo uqdactl getPeers
```

On a connected client, confirm it has both an ordinary private IPv4 address and
an UQDA IPv6 address from the gateway's `3xx:` subnet. Then ping a known UQDA
peer. A peer being `Up` on the gateway is not enough: the client test confirms
Router Advertisements and forwarding work end to end.

To remove only the configuration owned by this tool:

```sh
sudo uqda-gateway rollback
```

Keep local-console or Ethernet access during the first application. OpenWrt
rollback imports the configuration snapshots saved before the most recent
apply. NetworkManager rollback deletes the `uqda-gateway` connection and its
owned sysctl/radvd files; it does not modify unrelated connections.

## Security checklist

- use a unique 16+ character Wi-Fi password and WPA3 where every client supports it;
- keep OpenWrt/Linux and UQDA updated;
- do not expose the UQDA admin socket through TCP or Wi-Fi;
- firewall services on the gateway and clients;
- do not assume overlay encryption makes an application trustworthy; and
- test upgrades on a spare device before updating the household gateway.

The automation has deterministic tests, but Wi-Fi chips, drivers, regulatory
domains, and OpenWrt device packages vary. A release should not claim a router
model is validated until installation, reboot persistence, client SLAAC,
ordinary Internet access, UQDA peer reachability, and rollback have all been
tested on that model.
