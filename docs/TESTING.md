# Package Testing Guide

**Date:** January 16, 2026  
**Version:** Uqda Core v0.1.0

---

## Overview

This document describes the automated testing workflow for Uqda Core packages across all supported platforms: Linux (Debian/Ubuntu), Windows, and macOS.

---

## Automated Testing

### GitHub Actions Workflow

The `.github/workflows/test-packages.yml` workflow automatically tests all packages when:

- Code is pushed to `main` branch
- Pull requests are opened to `main`
- A release is published
- Manually triggered via `workflow_dispatch`

### Test Coverage

#### Linux (Debian/Ubuntu)

**Tested Architectures:**
- `amd64` (Intel/AMD 64-bit)
- `arm64` (ARM 64-bit)

**Test Steps:**
1. ✅ Build `.deb` package
2. ✅ Install package using `dpkg`
3. ✅ Verify binaries exist (`uqda`, `uqdactl`)
4. ✅ Test binary execution (`-version`, `-help`)
5. ✅ Check systemd service file exists
6. ✅ Test configuration generation (`-genconf`)
7. ✅ Test `uqdactl` commands
8. ✅ Test service status queries
9. ✅ Test package removal

**Example Output:**
```
✓ Binaries found and executable
✓ systemd service file exists
✓ Configuration generation works
✓ uqdactl commands work
✓ Service can be queried
✓ Package removal works
```

---

#### Windows

**Tested Architectures:**
- `x64` (Intel/AMD 64-bit)
- `arm64` (ARM 64-bit)

**Test Steps:**
1. ✅ Build `.msi` package
2. ✅ Install package using `msiexec`
3. ✅ Verify binaries exist in `C:\Program Files\Uqda\`
4. ✅ Test binary execution (`-version`, `-help`)
5. ✅ Test configuration generation (`-genconf`)
6. ✅ Test `uqdactl` commands
7. ✅ Test service status queries
8. ✅ Test package removal

**Example Output:**
```
✓ Installation completed
✓ Binaries found
✓ Configuration generation works
✓ uqdactl commands work
✓ Service can be queried
✓ Package removal works
```

---

#### macOS

**Tested Architectures:**
- `amd64` (Intel Macs)
- `arm64` (Apple Silicon - M1/M2/M3)

**Test Steps:**
1. ✅ Build `.pkg` package
2. ✅ Install package using `installer`
3. ✅ Verify binaries exist (`/usr/local/bin/uqda`, `/usr/local/bin/uqdactl`)
4. ✅ Test binary execution (`-version`, `-help`)
5. ✅ Check launchd plist file exists and is valid
6. ✅ Test configuration generation (`-genconf`)
7. ✅ Test `uqda -autoconf` command
8. ✅ Test `uqdactl` commands
9. ✅ Test launchd service queries
10. ✅ Check file permissions
11. ✅ Test architecture compatibility (Rosetta on Apple Silicon)
12. ✅ Test package removal

**Example Output:**
```
✓ Installation completed
✓ Binaries found and executable
✓ launchd plist file exists
✓ plist syntax is valid
✓ plist content is correct
✓ Configuration generation works
✓ autoconf command works
✓ uqdactl commands work
✓ launchd service can be queried
✓ File permissions are correct
✓ Testing native Apple Silicon package
✓ Package removal works
```

---

## Manual Testing

### Linux

```bash
# Build package
PKGARCH=amd64 sh contrib/deb/generate.sh

# Install
sudo dpkg -i *.deb

# Test
uqda -version
uqdactl -help
sudo uqda -genconf > /tmp/test.conf
sudo systemctl status uqda

# Remove
sudo dpkg -r uqda-v0.1.0
```

---

### Windows

```powershell
# Build package (requires Git Bash or WSL)
bash contrib/msi/build-msi.sh x64

# Install
msiexec /i *.msi /quiet

# Test
& "C:\Program Files\Uqda\uqda.exe" -version
& "C:\Program Files\Uqda\uqdactl.exe" -help
& "C:\Program Files\Uqda\uqda.exe" -genconf | Out-File C:\ProgramData\Uqda\test.conf

# Check service
Get-Service Uqda

# Remove
msiexec /x *.msi /quiet
```

---

### macOS

```bash
# Build package
PKGARCH=arm64 sh contrib/macos/create-pkg.sh

# Install
sudo installer -pkg *.pkg -target /

# Test
/usr/local/bin/uqda -version
/usr/local/bin/uqdactl -help
sudo /usr/local/bin/uqda -genconf > /tmp/test.conf
sudo /usr/local/bin/uqda -autoconf

# Check service
sudo launchctl list network.uqda.uqda
plutil -lint /Library/LaunchDaemons/uqda.plist

# Remove
sudo launchctl unload /Library/LaunchDaemons/uqda.plist
sudo rm -f /usr/local/bin/uqda /usr/local/bin/uqdactl
sudo rm -f /Library/LaunchDaemons/uqda.plist
```

---

## Troubleshooting

### Linux Tests Fail

**Problem:** `uqda: command not found`

**Solution:**
- Check if package installed correctly: `dpkg -l | grep uqda`
- Verify binaries: `dpkg -L uqda-v0.1.0`
- Check PATH: `which uqda`

---

**Problem:** `systemd service file not found`

**Solution:**
- Check service file location:
  ```bash
  ls -la /etc/systemd/system/uqda.service
  ls -la /lib/systemd/system/uqda.service
  ```
- Verify package contents: `dpkg -L uqda-v0.1.0 | grep service`

---

### Windows Tests Fail

**Problem:** `MSI installation fails`

**Solution:**
- Check installation log: `Get-Content install.log`
- Verify .NET SDK is installed (required for MSI builder)
- Check file paths in MSI: `msiexec /i *.msi /l*v install.log`

---

**Problem:** `Binaries not found`

**Solution:**
- Check installation path: `Test-Path "C:\Program Files\Uqda\uqda.exe"`
- Verify MSI contents: `msiexec /i *.msi /l*v install.log`
- Check for installation errors in log

---

### macOS Tests Fail

**Problem:** `plist file not found`

**Solution:**
- Check installation: `ls -la /Library/LaunchDaemons/uqda.plist`
- Verify package contents: `pkgutil --files io.github.Uqda-network.pkg`
- Check postinstall script executed correctly

---

**Problem:** `launchd service won't load`

**Solution:**
- Check plist syntax: `plutil -lint /Library/LaunchDaemons/uqda.plist`
- Check service label: `grep -A 1 Label /Library/LaunchDaemons/uqda.plist`
- Check logs: `log show --predicate 'process == "uqda"' --last 5m`

---

**Problem:** `Architecture mismatch`

**Solution:**
- On Apple Silicon, Intel packages work via Rosetta
- Verify architecture: `uname -m` (should be `arm64`)
- Check package architecture: `file /usr/local/bin/uqda`

---

## Continuous Integration

### Running Tests Locally

You can simulate GitHub Actions locally using [act](https://github.com/nektos/act):

```bash
# Install act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run Linux tests
act -j test-linux-debian

# Run Windows tests (requires Docker)
act -j test-windows

# Run macOS tests (requires macOS runner)
act -j test-macos
```

---

## Test Results

All tests must pass before:
- Merging pull requests
- Creating releases
- Publishing packages

**Success Criteria:**
- ✅ All packages build successfully
- ✅ All packages install correctly
- ✅ All binaries execute properly
- ✅ All services can be queried
- ✅ All packages can be removed cleanly

---

## Reporting Issues

If tests fail, please report:

1. **Platform:** Linux/Windows/macOS
2. **Architecture:** amd64/arm64/x64
3. **Test Step:** Which step failed
4. **Error Message:** Full error output
5. **Logs:** Relevant log files

**GitHub Issues:** https://github.com/Uqda/Core/issues

---

## References

- **GitHub Actions:** https://docs.github.com/en/actions
- **Uqda Core Repository:** https://github.com/Uqda/Core
- **Testing Workflow:** `.github/workflows/test-packages.yml`

---

**Last Updated:** January 16, 2026  
**Status:** Active

