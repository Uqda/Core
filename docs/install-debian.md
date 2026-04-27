# Installing on Debian, Ubuntu, elementaryOS, Linux Mint and similar

These instructions apply to **Debian-based** distributions using **systemd**.

## One-off installation from package file

Debian packages may be attached to [GitHub Releases](https://github.com/Uqda/Core/releases). Install with:

```sh
sudo dpkg -i uqda_x.x.x_amd64.deb
```

Resolve dependencies if **`dpkg`** complains:

```sh
sudo apt-get install -f
```

## Package repository (optional)

**(Uqda package repository — see [https://github.com/Uqda/Core/releases](https://github.com/Uqda/Core/releases))**

If you maintain an apt repository for Uqda, replace the repository URL and signing key with your own. Until then, prefer **`dpkg -i`** from releases or build from source.

## Distribution packages

Some distributions may ship an **`uqda`** or forked package:

```sh
sudo apt-get install uqda
```

**Warning:** Distribution packages are often **out of date**. If you cannot peer or see protocol mismatches, compare **`apt show uqda`** with the latest release on GitHub.

## After installation

Configuration is commonly **`/etc/uqda.conf`**. Enable and start:

```sh
sudo systemctl enable uqda
sudo systemctl start uqda
```

After edits:

```sh
sudo systemctl restart uqda
```

## See also

- [Uninstall completely](uninstall.md)
