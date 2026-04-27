# Installing using the macOS installer

Uqda is supported on macOS. Download the **`.pkg`** from [GitHub Releases](https://github.com/Uqda/Core/releases).

## Install using Finder

Open the downloaded `.pkg` (right-click → **Open** if Gatekeeper warns), then follow the installer.

When complete, configuration is typically generated, **launchd** loads **`/Library/LaunchDaemons/uqda.plist`**, and the service runs.

## Install using Terminal

```sh
sudo installer -pkg /path/to/uqda-xxx-macos.pkg -target /
```

## Managing the service

```sh
sudo launchctl unload /Library/LaunchDaemons/uqda.plist
sudo launchctl load /Library/LaunchDaemons/uqda.plist
```

Service name and labels follow the plist’s `Label` (for example **`uqda`**).

## See also

- [Uninstall completely](uninstall.md)
