# Installation Guide - Uqda Network

## System Requirements

- **Go:** 1.22 or later
- **Git:** For cloning the repository
- **Root/sudo privileges:** To create TUN/TAP interface

## Installation from Go

### Quick Method

```bash
go install github.com/Uqda/Core/cmd/uqda@latest
go install github.com/Uqda/Core/cmd/uqdactl@latest
```

### Verify Installation

```bash
uqda -version
uqdactl -version
```

## Building from Source

### Linux (Ubuntu/Debian)

```bash
# Install dependencies
sudo apt update
sudo apt install -y golang-go git build-essential

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/

# Verify
uqda -version
```

### Linux (Fedora/RHEL/CentOS)

```bash
# Install dependencies
sudo dnf install -y golang git gcc

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### Linux (Arch Linux)

```bash
# Install dependencies
sudo pacman -S go git base-devel

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### macOS

```bash
# Install Go (if not installed)
brew install go git

# Or from golang.org
# https://golang.org/dl/

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### Windows

#### Using PowerShell

```powershell
# Install Go from golang.org
# https://golang.org/dl/

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
go build -o uqda.exe ./cmd/uqda
go build -o uqdactl.exe ./cmd/uqdactl

# Executables will be in the same directory
```

#### Using WSL (Windows Subsystem for Linux)

```bash
# Follow Linux instructions above inside WSL
```

### FreeBSD

```bash
# Install Go
sudo pkg install go git

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### OpenBSD

```bash
# Install Go
doas pkg_add go git

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
doas cp uqda /usr/local/bin/
doas cp uqdactl /usr/local/bin/
```

### Android (Termux)

```bash
# Install dependencies
pkg install golang git

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
cp uqda ~/../usr/bin/
cp uqdactl ~/../usr/bin/
```

### OpenWrt

```bash
# Install dependencies
opkg update
opkg install golang git

# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Copy executables
cp uqda /usr/bin/
cp uqdactl /usr/bin/
```

## Installation as a Service (Linux - systemd)

### Create Service File

```bash
sudo nano /etc/systemd/system/uqda.service
```

```ini
[Unit]
Description=Uqda Network Router
Documentation=https://github.com/Uqda/Core
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/uqda -useconffile /etc/uqda/uqda.conf
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Enable Service

```bash
# Create config directory
sudo mkdir -p /etc/uqda

# Generate config
sudo uqda -genconf > /etc/uqda/uqda.conf

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable uqda
sudo systemctl start uqda

# Check status
sudo systemctl status uqda
```

## Installation on macOS (LaunchDaemon)

### Create LaunchDaemon File

```bash
sudo nano /Library/LaunchDaemons/network.uqda.uqda.plist
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>network.uqda.uqda</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/uqda</string>
      <string>-useconffile</string>
      <string>/etc/uqda/uqda.conf</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
  </dict>
</plist>
```

### Enable Service

```bash
# Create config directory
sudo mkdir -p /etc/uqda

# Generate config
sudo uqda -genconf > /etc/uqda/uqda.conf

# Load service
sudo launchctl load /Library/LaunchDaemons/network.uqda.uqda.plist
```

## Installation on Windows (Service)

### Using NSSM (Non-Sucking Service Manager)

```powershell
# Download NSSM from nssm.cc

# Install service
nssm install Uqda "C:\path\to\uqda.exe" "-useconffile C:\path\to\uqda.conf"

# Start service
nssm start Uqda
```

## Cross-Compilation (Building for Other Systems)

### Build for Windows from Linux

```bash
GOOS=windows GOARCH=amd64 ./build
```

### Build for macOS from Linux

```bash
GOOS=darwin GOARCH=amd64 ./build
```

### Build for ARM (Raspberry Pi)

```bash
GOOS=linux GOARCH=arm GOARM=7 ./build
```

### Build for MIPS (OpenWrt)

```bash
GOOS=linux GOARCH=mipsle ./build
```

## Verify Installation

```bash
# Show version information
uqda -version

# Show node information (after running)
uqdactl getSelf

# Show connected peers
uqdactl getPeers
```

## Troubleshooting

### Issue: "Permission denied" when creating TUN

**Solution on Linux:**

```bash
# Grant CAP_NET_ADMIN capability
sudo setcap cap_net_admin+eip /usr/local/bin/uqda
```

### Issue: "Command not found"

**Solution:**

```bash
# Check if files are in PATH
which uqda
which uqdactl

# Add /usr/local/bin to PATH if needed
export PATH=$PATH:/usr/local/bin
```

### Issue: "Cannot create TUN interface"

**Solution:**

```bash
# On Linux: Load TUN module
sudo modprobe tun

# On macOS: May need admin privileges
sudo uqda -useconffile /etc/uqda/uqda.conf
```

---

[Previous: Introduction ←](01-introduction.md) | [Next: Quick Start →](03-quickstart.md)

