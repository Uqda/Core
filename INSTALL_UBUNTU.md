# Installation Guide for Ubuntu Server

Complete installation guide for Uqda Network on Ubuntu Server.

## Prerequisites

- Ubuntu Server 18.04 LTS or later (20.04, 22.04, 24.04 recommended)
- Root or sudo access
- Internet connection

## Method 1: Install from Pre-built Package (Recommended)

### Step 1: Download the Package

```bash
# For 64-bit systems (amd64)
wget https://github.com/Uqda/Core/releases/download/v0.1.0/Uqda-v0.1.0-0.1.0-amd64.deb

# For ARM64 systems
# wget https://github.com/Uqda/Core/releases/download/v0.1.0/Uqda-v0.1.0-0.1.0-arm64.deb
```

### Step 2: Install the Package

```bash
sudo dpkg -i Uqda-v0.1.0-0.1.0-amd64.deb

# If you get dependency errors, fix them with:
sudo apt-get install -f
```

### Step 3: Verify Installation

```bash
# Check if binaries are installed
which uqda
which uqdactl

# Check version
uqda -version
uqdactl -version
```

### Step 4: Check Service Status

```bash
# Check if service is running
sudo systemctl status uqda

# If not running, start it
sudo systemctl start uqda
sudo systemctl enable uqda
```

### Step 5: View Your Node Information

```bash
# Get your node information
sudo uqdactl getSelf

# View connected peers
sudo uqdactl getPeers
```

---

## Method 2: Build from Source

### Step 1: Install Dependencies

```bash
# Update package list
sudo apt-get update

# Install Go (1.22 or later)
sudo apt-get install -y golang-go git build-essential

# Verify Go installation
go version
```

### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build for your system
./build

# Verify binaries were created
ls -lh uqda uqdactl
```

### Step 3: Install Binaries

```bash
# Copy binaries to system path
sudo cp uqda uqdactl /usr/local/bin/

# Make them executable (if needed)
sudo chmod +x /usr/local/bin/uqda
sudo chmod +x /usr/local/bin/uqdactl

# Verify installation
which uqda
which uqdactl
```

### Step 4: Create Configuration Directory

```bash
# Create config directory
sudo mkdir -p /etc/uqda

# Generate configuration
sudo uqda -genconf > /etc/uqda/uqda.conf

# Set proper permissions
sudo chown root:root /etc/uqda/uqda.conf
sudo chmod 644 /etc/uqda/uqda.conf
```

### Step 5: Create Systemd Service

```bash
# Create systemd service file
sudo nano /etc/systemd/system/uqda.service
```

Add the following content:

```ini
[Unit]
Description=Uqda Network
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/uqda -useconffile /etc/uqda/uqda.conf
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Save and exit (Ctrl+X, then Y, then Enter).

### Step 6: Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable uqda

# Start the service
sudo systemctl start uqda

# Check status
sudo systemctl status uqda
```

---

## Configuration

### Basic Configuration

The configuration file is located at `/etc/uqda/uqda.conf`.

```bash
# Edit configuration
sudo nano /etc/uqda/uqda.conf
```

### Add Peers

You can add peers in two ways:

**Method 1: Edit configuration file**

```bash
sudo nano /etc/uqda/uqda.conf
```

Add peers to the `Peers` array:

```json
{
  "Peers": [
    "tcp://example.com:12345",
    "tls://secure-peer.example.com:8080"
  ]
}
```

Then restart the service:

```bash
sudo systemctl restart uqda
```

**Method 2: Use uqdactl (while service is running)**

```bash
# Add a peer
sudo uqdactl addPeer tcp://example.com:12345

# View peers
sudo uqdactl getPeers
```

### Auto-Configuration Mode

If you want to run without a persistent configuration:

```bash
# Stop the service
sudo systemctl stop uqda

# Edit service file
sudo systemctl edit uqda
```

Add:

```ini
[Service]
ExecStart=
ExecStart=/usr/local/bin/uqda -autoconf
```

Then restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart uqda
```

---

## Useful Commands

### Service Management

```bash
# Start service
sudo systemctl start uqda

# Stop service
sudo systemctl stop uqda

# Restart service
sudo systemctl restart uqda

# Check status
sudo systemctl status uqda

# View logs
sudo journalctl -u uqda -f
```

### Node Information

```bash
# Get your node information
sudo uqdactl getSelf

# View connected peers
sudo uqdactl getPeers

# View routing tree
sudo uqdactl getTree

# View active sessions
sudo uqdactl getSessions

# View routing paths
sudo uqdactl getPaths
```

### Peer Management

```bash
# Add a peer
sudo uqdactl addPeer tcp://peer.example.com:12345

# Remove a peer
sudo uqdactl removePeer tcp://peer.example.com:12345

# List all available commands
sudo uqdactl list
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check service status
sudo systemctl status uqda

# View detailed logs
sudo journalctl -u uqda -n 50

# Check if configuration file exists
ls -la /etc/uqda/uqda.conf

# Test configuration
sudo uqda -useconffile /etc/uqda/uqda.conf -normaliseconf
```

### Can't Connect to Peers

```bash
# Check if peers are configured
sudo uqdactl getPeers

# Check firewall
sudo ufw status

# If firewall is active, allow Uqda ports
sudo ufw allow 12345/tcp
```

### Permission Issues

```bash
# Check file permissions
ls -la /usr/local/bin/uqda
ls -la /etc/uqda/uqda.conf

# Fix permissions if needed
sudo chmod +x /usr/local/bin/uqda
sudo chmod 644 /etc/uqda/uqda.conf
```

### Network Interface Issues

```bash
# Check if TUN interface is created
ip addr show | grep tun

# Load TUN module if needed
sudo modprobe tun

# Check if module is loaded
lsmod | grep tun
```

---

## Firewall Configuration

If you're running a firewall (UFW), you may need to allow incoming connections:

```bash
# Check firewall status
sudo ufw status

# Allow specific port (if you're listening on a specific port)
sudo ufw allow 12345/tcp

# Or allow all Uqda traffic (less secure)
sudo ufw allow from any to any port 12345
```

---

## Uninstallation

### If Installed from Package

```bash
# Remove package
sudo dpkg -r uqda-v0.1.0

# Remove configuration (optional)
sudo rm -rf /etc/uqda
```

### If Installed from Source

```bash
# Stop and disable service
sudo systemctl stop uqda
sudo systemctl disable uqda

# Remove service file
sudo rm /etc/systemd/system/uqda.service
sudo systemctl daemon-reload

# Remove binaries
sudo rm /usr/local/bin/uqda
sudo rm /usr/local/bin/uqdactl

# Remove configuration (optional)
sudo rm -rf /etc/uqda
```

---

## Quick Start Summary

For a quick installation:

```bash
# Download and install
wget https://github.com/Uqda/Core/releases/download/v0.1.0/Uqda-v0.1.0-0.1.0-amd64.deb
sudo dpkg -i Uqda-v0.1.0-0.1.0-amd64.deb
sudo apt-get install -f

# Check status
sudo systemctl status uqda

# Get your node info
sudo uqdactl getSelf

# Add a peer
sudo uqdactl addPeer tcp://peer.example.com:12345
```

---

## Additional Resources

- **GitHub Repository**: https://github.com/Uqda/Core
- **Issues**: https://github.com/Uqda/Core/issues
- **Discussions**: https://github.com/Uqda/Core/discussions
- **Email Support**: uqda@proton.me

---

**Note**: Make sure to replace example peer addresses with actual peer addresses from the Uqda network.

