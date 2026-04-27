# Installing on Gentoo Linux

Uqda may be published in the **GURU** overlay as **`net-p2p/uqda`** (package name may vary while the overlay is updated—search **`net-p2p/uqda`** or **`uqda`** in GURU).

## Enable GURU

```sh
sudo eselect repository enable guru
sudo emerge --sync
```

## Keywords

If the package is **`~amd64`** (or similar), accept the keyword:

If **`/etc/portage/package.accept_keywords`** is a directory:

```sh
echo "net-p2p/uqda ~amd64" | sudo tee /etc/portage/package.accept_keywords/uqda
```

If it is a file:

```sh
echo "net-p2p/uqda ~amd64" | sudo tee -a /etc/portage/package.accept_keywords
```

## Install

```sh
sudo emerge --ask net-p2p/uqda
```

## Configuration

Edit **`/etc/uqda.conf`**, then reload or restart:

**OpenRC**

```sh
rc-service uqda reload
# or
rc-service uqda restart
```

**systemd**

```sh
systemctl reload uqda
# or
systemctl restart uqda
```
