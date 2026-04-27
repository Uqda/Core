# Installing on Red Hat Enterprise Linux, Fedora, CentOS and similar

Applies to **RPM-based** distributions with **systemd**.

## COPR / third-party repository

Fedora COPR and similar repositories change over time. See **[https://github.com/Uqda/Core](https://github.com/Uqda/Core)** for current packaging notes or COPR links if maintained.

Example once a repository is enabled:

```sh
sudo dnf install uqda
```

## Configuration

Edit **`/etc/uqda.conf`**, then:

```sh
sudo systemctl restart uqda
```

## Linux documentation

See **[install-linux-manual.md](install-linux-manual.md)** for building from source on RPM-based systems without a prebuilt package.
