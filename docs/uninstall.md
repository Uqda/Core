# Uninstall Uqda completely

This guide removes **Uqda**, **uqdactl**, service entries, and typical configuration files so you can reinstall from a clean state.

**Warning:** Deleting configuration **destroys your private key** and therefore your **Uqda IPv6 address** on the network. Back up `uqda.conf` (or export the key) before removal if you need to keep the same identity.

---

## macOS

### If you used **launchd** (manual or `.pkg`)

```sh
sudo launchctl unload /Library/LaunchDaemons/uqda.plist 2>/dev/null || true
sudo rm -f /Library/LaunchDaemons/uqda.plist
sudo rm -f /usr/local/bin/uqda /usr/local/bin/uqdactl
sudo rm -f /etc/uqda.conf
sudo rm -f /var/run/uqda.sock
```

Remove log files if you used the sample plist paths:

```sh
sudo rm -f /tmp/uqda.stdout.log /tmp/uqda.stderr.log
```

### If you only built from a **git checkout**

Stop any running `uqda` process, then delete the repository folder (or remove only `uqda` / `uqdactl` binaries you copied).

---

## Linux (systemd, typical paths)

```sh
sudo systemctl stop uqda 2>/dev/null || true
sudo systemctl disable uqda 2>/dev/null || true
sudo rm -f /etc/systemd/system/uqda.service
sudo systemctl daemon-reload
```

Binaries and config (adjust paths if you installed elsewhere):

```sh
sudo rm -f /usr/local/bin/uqda /usr/local/bin/uqdactl
sudo rm -f /usr/sbin/uqda /usr/sbin/uqdactl 2>/dev/null || true
sudo rm -f /etc/uqda.conf
sudo rm -rf /etc/uqda
sudo rm -f /var/run/uqda.sock
```

### Debian / Ubuntu package

```sh
sudo apt-get remove --purge uqda
```

Then remove any leftover config you created manually under `/etc/` if still present.

### RPM-based

```sh
sudo dnf remove uqda 2>/dev/null || sudo yum remove uqda 2>/dev/null || true
```

### OpenRC (Gentoo / Alpine style)

```sh
sudo rc-service uqda stop 2>/dev/null || true
sudo rc-update del uqda default 2>/dev/null || true
# Remove init script path your distro used, e.g.:
sudo rm -f /etc/init.d/uqda
```

---

## Windows

1. **Settings → Apps → Installed apps** → find **Uqda** → **Uninstall** (or use the MSI again with remove/repair if offered).
2. Delete configuration directory if it remains:

`%ALLUSERSPROFILE%\Uqda\`

(Explorer: paste into the address bar.)

3. Open **Services** (`services.msc`) and confirm no **Uqda** service remains.
4. Optional: **Device Manager** → **Network adapters** → remove a stale **Uqda** / Wintun adapter only if still listed after uninstall (reboot first).

---

## Ubiquiti EdgeOS / VyOS (vyatta-uqda)

Remove the interface from configuration (replace `tun0` with yours):

```
configure
delete interfaces uqda tun0
commit
save
```

Remove package and per-interface config on disk:

```sh
sudo dpkg -r vyatta-uqda   # or: dpkg -l | grep -i uqda  to see exact name
sudo rm -f /config/uqda.tun0.conf
```

Reboot or `restart` relevant services per your platform documentation if something still references the old interface.

---

## FreeBSD

### From **pkg**

```sh
sudo pkg delete uqda
```

### Manual / rc.d script

```sh
sudo service uqda onestop 2>/dev/null || true
sudo sysrc -x uqda_enable 2>/dev/null || true
sudo rm -f /usr/local/etc/rc.d/uqda
sudo rm -f /usr/local/sbin/uqda /usr/local/sbin/uqdactl
sudo rm -f /usr/local/etc/uqda.conf
```

---

## OpenWrt

```sh
opkg remove luci-proto-uqda 2>/dev/null || true
opkg remove uqda
rm -f /etc/config/uqda
/etc/init.d/network reload
```

(Exact package names may match your feed.)

---

## After removal checklist

- [ ] No `uqda` / `uqdactl` process: `ps aux | grep uqda` (Linux/macOS) or Task Manager (Windows).
- [ ] No admin socket: e.g. `/var/run/uqda.sock` removed or unused.
- [ ] No stale **TUN** / virtual adapter causing routing confusion (reboot once).
- [ ] Firewall rules you added **only** for Uqda — remove if no longer needed.

---

## Reinstall

Follow the [installation guides](install-linux-manual.md) for your platform (index in the repository [README](../README.md#documentation)).
