# Installing manually on Linux

Uqda is supported on Linux.

## Prerequisites

Install **Go** from your distribution or [https://go.dev/dl/](https://go.dev/dl/). Match the `go` directive in **`go.mod`**.

## Build from source

```sh
cd /path/to
git clone https://github.com/Uqda/Core
cd Core
./build
```

Install binaries:

```sh
sudo cp uqda uqdactl /usr/local/bin
```

### Debug builds

```sh
./build -d
```

## systemd service

Service files live under **`contrib/systemd/`** (for example **`uqda.service`**) and expect configuration at **`/etc/uqda.conf`**.

Create a dedicated group:

```sh
sudo groupadd --system uqda
```

Install the unit file:

```sh
sudo cp contrib/systemd/uqda.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable uqda
sudo systemctl start uqda
```

View logs:

```sh
systemctl status uqda
journalctl -u uqda
```

## Generate configuration

```sh
sudo uqda -genconf > /etc/uqda.conf
```

Edit **`/etc/uqda.conf`**, then **`systemctl restart uqda`**.

## See also

- [Uninstall completely](uninstall.md)
- General **Linux** documentation in the Uqda project wiki or [https://uqda-network.github.io/](https://uqda-network.github.io/) when published.
