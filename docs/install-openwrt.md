# Installing on OpenWrt

Uqda has been packaged for OpenWrt in community trees; availability depends on your OpenWrt version and feeds.

Install the **`uqda`** package (and **`luci-proto-uqda`** for LuCI integration when available).

## LuCI

Use **Network** → interfaces → add **Uqda** when **`luci-proto-uqda`** is installed.

## Command line

Show options:

```sh
uci show uqda
```

Add a peer:

```sh
uci add uqda peer
uci set uqda.@peer[-1].uri='tcp://1.2.3.4:5678'
uci commit
/etc/init.d/uqda restart
```

Configuration file: **`/etc/config/uqda`**.

## Community

Matrix: **`#uqda-openwrt:matrix.org`**

## NodeInfo

The daemon may expose basic **NodeInfo** to the mesh by default (kernel, hostname, model, board name). Disable this in **`/etc/config/uqda`** or LuCI if you prefer less metadata.
