# Development and Contributing - Uqda Network

## Technical Architecture

### Language and Dependencies

**Language:** Go (Golang) 1.22+

**Main Dependencies:**

- `golang.org/x/crypto` - Cryptography
- `golang.org/x/net` - Networking
- `github.com/hjson/hjson-go` - HJSON parsing
- `github.com/Arceliar/ironwood` - Mesh networking
- `golang.zx2c4.com/wireguard` - TUN interface

### Code Structure

```
src/
├── core/           # Core protocol logic
│   ├── core.go     # Main core type
│   ├── link.go     # Peering connections
│   ├── router.go   # Routing logic
│   └── proto.go    # Protocol handlers
├── tun/            # TUN/TAP interface
├── multicast/      # Multicast discovery
├── config/         # Configuration handling
├── admin/          # Admin API
└── address/        # Address derivation
```

## Building from Source

### Requirements

```bash
# Go 1.22 or later
go version

# Git
git --version
```

### Build

```bash
# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build
./build

# Or manually
go build -o uqda ./cmd/uqda
go build -o uqdactl ./cmd/uqdactl
```

### Cross-Compilation

```bash
# Windows
GOOS=windows GOARCH=amd64 ./build

# macOS
GOOS=darwin GOARCH=amd64 ./build

# ARM (Raspberry Pi)
GOOS=linux GOARCH=arm GOARM=7 ./build

# MIPS (OpenWrt)
GOOS=linux GOARCH=mipsle ./build
```

## Testing

### Run Tests

```bash
# All tests
go test ./...

# Specific tests
go test ./src/core/...

# With verbose
go test -v ./...

# With coverage
go test -cover ./...
```

### Integration Tests

```bash
# Network tests
cd misc
./run-twolink-test
./run-schannel-netns
```

## Contributing

### How to Contribute

1. **Fork the repository**
2. **Create a new branch**
3. **Make changes**
4. **Add tests**
5. **Submit Pull Request**

### Code Standards

```bash
# Format code
go fmt ./...

# Check for errors
go vet ./...

# golangci-lint
golangci-lint run
```

### Commit Messages

```
type(scope): subject

body

footer
```

**Examples:**

```
feat(core): add new routing algorithm

fix(config): resolve BOM encoding issue

docs(readme): update installation instructions
```

## Project Structure

### Main Directories

```
Uqda/Core/
├── cmd/              # Executables
│   ├── uqda/         # Main program
│   └── uqdactl/      # Control tool
├── src/              # Source code
│   ├── core/         # Core protocol
│   ├── config/       # Configuration
│   ├── tun/          # TUN/TAP
│   └── multicast/    # Discovery
├── contrib/          # Additional contributions
│   ├── systemd/      # systemd files
│   ├── docker/       # Docker
│   └── ansible/      # Ansible
└── misc/             # Additional tools
```

## Internal API

### Core API

```go
// Create new core
core, err := core.New(certificate, logger, options...)

// Get address
address := core.Address()

// Get subnet
subnet := core.Subnet()

// Stop core
core.Stop()
```

### Admin API

```go
// Create admin socket
admin, err := admin.New(core, logger, options...)

// Setup handlers
admin.SetupAdminHandlers()
```

## Future Development

### Planned Features

- [ ] Performance improvements
- [ ] Additional security features
- [ ] Support for new platforms
- [ ] Routing improvements

### How to Contribute

1. **Pick an issue from GitHub**
2. **Create a new branch**
3. **Write code**
4. **Add tests**
5. **Submit PR**

---

[Previous: Troubleshooting ←](08-troubleshooting.md) | [Back to Index →](README.md)

