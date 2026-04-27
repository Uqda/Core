# Installing on FreeBSD

> **Warning:** FreeBSD TUN adapter support is currently incomplete and may not work correctly. Use Uqda on FreeBSD for routing/peering only, without a TUN interface (set `IfName: none` in your config), until this is resolved.

Uqda can be installed from **pkg** once a port or package is available, or built from source.

## Binary package (future)

When **`uqda`** appears in the ports tree or official packages:

```sh
sudo pkg install uqda
```

Until then, build from ports (when committed) or from source below.

## Configuration path

Typical locations:

- **`/usr/local/etc/uqda.conf`**

Generate:

```sh
sudo uqda -genconf > /usr/local/etc/uqda.conf
```

## RC service

Enable and start (names follow the **`uqda`** rc script):

```sh
sudo sysrc uqda_enable=YES
sudo service uqda start
```

Or:

```sh
sudo service uqda enable
sudo service uqda start
```

(FreeBSD version-dependent; **`sysrc`** + **`service`** are common.)

## Build from source

```sh
git clone https://github.com/Uqda/Core
cd Core
./build
sudo cp uqda uqdactl /usr/local/sbin/
```

See **`contrib/freebsd/`** and **`contrib/freebsd-port/`** in this repository for rc script and port skeleton.

## Related

See **`contrib/freebsd-port/`** in this repository for a ports-tree skeleton you can merge or adapt locally.
