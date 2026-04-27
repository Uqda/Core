# Installing manually on macOS

Uqda is supported on macOS.

## Prerequisites

Install the **Go** toolchain ([https://go.dev/dl/](https://go.dev/dl/)). Uqda tracks a recent Go release; see `go.mod` in the repository for the minimum version.

## Build from source

```sh
cd /path/to
git clone https://github.com/Uqda/Core
cd Core
./build
```

This produces **`uqda`** and **`uqdactl`** in the repository root.

System Integrity Protection prevents copying into `/usr/bin`; install into **`/usr/local/bin`**:

```sh
sudo cp uqda uqdactl /usr/local/bin
```

### Debug builds

```sh
./build -d
```

Debug builds include extra diagnostics and may be larger.

## launchd service

Example plist for background operation is in **`contrib/macos/uqda.plist`** (configuration path **`/etc/uqda.conf`**).

```sh
sudo cp contrib/macos/uqda.plist /Library/LaunchDaemons/
sudo launchctl load /Library/LaunchDaemons/uqda.plist
```

Logs (if configured as in the sample plist):

```sh
tail -f /tmp/uqda.stdout.log
tail -f /tmp/uqda.stderr.log
```

## Generate configuration

```sh
sudo uqda -genconf > /etc/uqda.conf
```

Edit the file as needed, then restart the service or run **Uqda** with `-useconffile /etc/uqda.conf`.

## See also

- [Uninstall completely](uninstall.md)
