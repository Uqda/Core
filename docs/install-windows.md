# Installing using the Windows installer

Uqda is supported on Windows on a **best-effort** basis only. Download the latest installer from [GitHub Releases](https://github.com/Uqda/Core/releases).

## Windows 7 / Server 2008 R2

You must install hotfix **KB2921916** before installing Uqda:

- [KB2921916 for 64-bit systems](https://support.microsoft.com/en-us/kb/2921916)
- [KB2921916 for 32-bit systems](https://support.microsoft.com/en-us/kb/2921916)

## Warning

The Windows port does not currently have a dedicated maintainer and may be less well tested than other platforms.

## TUN driver

Uqda on Windows uses the **WireGuard TUN** driver (Wintun). If it is not installed, the MSI installer typically installs it. Use the installer that matches your architecture (**x64** on 64-bit Windows, **x86** on 32-bit).

The OpenVPN TAP driver is not used for current Uqda builds.

Once Uqda is started, a new virtual network adapter is created named **Uqda** by default (this matches the default **IfName** in the generated configuration). The adapter is not visible when Uqda is not running.

## Configuration

The installer can generate **`%ALLUSERSPROFILE%\Uqda\uqda.conf`** if it does not exist.

## Windows service

Uqda is installed as a Windows service that can start automatically. Use **services.msc** or the Services tab in Task Manager. Restart the service after each configuration file change.

## Windows Firewall

You may be prompted to allow **Uqda** through the firewall for peerings to work. You can also treat the Uqda adapter as a **public** network so unexpected incoming connections are limited; ensure SMB/RPC/RDP are not exposed on public networks unless intended.

## uqdactl

The **uqdactl** utility is installed alongside **Uqda**. Example from **Command Prompt** or **PowerShell**:

```bat
"C:\Program Files\Uqda\uqdactl.exe" getPeers
```

Adjust the path if you installed to a different directory.

## See also

- [Uninstall completely](uninstall.md)
