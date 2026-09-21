// Package config parses and generates Uqda node configuration. Persistent
// configurations must retain their private key to preserve node identity.
package config

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/hjson/hjson-go/v4"
	"golang.org/x/text/encoding/unicode"
)

// NodeConfig is the main configuration structure, containing configuration
// options that are necessary for a Uqda node to run. You will need to
// supply one of these structs to Uqda Core when starting a node.
type NodeConfig struct {
	PrivateKey          KeyBytes                   `json:",omitempty" comment:"Your private key. DO NOT share this with anyone!"`
	PrivateKeyPath      string                     `json:",omitempty" comment:"The path to your private key file in PEM format."`
	Certificate         *tls.Certificate           `json:"-"`
	Peers               []string                   `comment:"List of outbound peer connection strings (e.g. tls://a.b.c.d:e or\nsocks://a.b.c.d:e/f.g.h.i:j). Connection strings can contain options,\nsee https://yggdrasil-network.github.io/configurationref.html#peers.\nYggdrasil has no concept of bootstrap nodes - all network traffic\nwill transit peer connections. Therefore make sure to only peer with\nnearby nodes that have good connectivity and low latency. Avoid adding\npeers to this list from distant countries as this will worsen your\nnode's connectivity and performance considerably."`
	InterfacePeers      map[string][]string        `comment:"List of connection strings for outbound peer connections in URI format,\narranged by source interface, e.g. { \"eth0\": [ \"tls://a.b.c.d:e\" ] }.\nYou should only use this option if your machine is multi-homed and you\nwant to establish outbound peer connections on different interfaces.\nOtherwise you should use \"Peers\"."`
	Listen              []string                   `comment:"Listen addresses for incoming connections. You will need to add\nlisteners in order to accept incoming peerings from non-local nodes.\nThis is not required if you wish to establish outbound peerings only.\nMulticast peer discovery will work regardless of any listeners set\nhere. Each listener should be specified in URI format as above, e.g.\ntls://0.0.0.0:0 or tls://[::]:0 to listen on all interfaces."`
	AdminListen         string                     `json:",omitempty" comment:"Listen address for admin connections. Default is to listen for local\nconnections either on TCP/9001 or a UNIX socket depending on your\nplatform. Use this value for uqdactl -endpoint=X. To disable\nthe admin socket, use the value \"none\" instead."`
	MulticastInterfaces []MulticastInterfaceConfig `comment:"Configuration for which interfaces multicast peer discovery should be\nenabled on. Regex is a regular expression which is matched against an\ninterface name, and interfaces use the first configuration that they\nmatch against. Beacon controls whether or not your node advertises its\npresence to others, whereas Listen controls whether or not your node\nlistens out for and tries to connect to other advertising nodes. See\nhttps://yggdrasil-network.github.io/configurationref.html#multicastinterfaces\nfor more supported options."`
	AllowedPublicKeys   []string                   `comment:"List of peer public keys to allow incoming peering connections\nfrom. If left empty/undefined then all connections will be allowed\nby default. This does not affect outgoing peerings, nor does it\naffect link-local peers discovered via multicast.\nWARNING: THIS IS NOT A FIREWALL and DOES NOT limit who can reach\nopen ports or services running on your machine, for that see the\nGroupPassword option below."`
	GroupPassword       string                     `comment:"Traffic is only allowed to/from nodes with the same group password.\nIf you want to form a private sub-network or ensure that other public\nusers cannot connect to your machines, choose a strong group password\nand then configure the same password only with other group members.\nIf left empty or not specified, public connectivity will be permitted.\nIf specified, you WILL NOT be able to reach public services or hosts.\nThis option DOES NOT affect peering connections or traffic routing."`
	IfName              string                     `comment:"Local network interface name for TUN adapter, or \"auto\" to select\nan interface automatically, or \"none\" to run without TUN."`
	IfMTU               uint64                     `comment:"Maximum Transmission Unit (MTU) size for your local TUN interface.\nDefault is the largest supported size for your platform. The lowest\npossible value is 1280."`
	LogLookups          bool                       `json:",omitempty"`
	NodeInfoPrivacy     bool                       `comment:"By default, nodeinfo contains some defaults including the platform,\narchitecture and Uqda build version. These can help when surveying\nthe network and diagnosing network routing problems. Enabling\nnodeinfo privacy prevents this, so that only items specified in\n\"NodeInfo\" are sent back if specified."`
	NodeInfo            map[string]interface{}     `comment:"Optional nodeinfo. This must be a { \"key\": \"value\", ... } map\nor set as null. This is entirely optional but, if set, is visible\nto the whole network on request."`
}

// MulticastInterfaceConfig controls discovery on matching network interfaces.
type MulticastInterfaceConfig struct {
	Regex    string
	Beacon   bool
	Listen   bool
	Port     uint16 `json:",omitempty"`
	Priority uint64 `json:",omitempty"` // really uint8, but gobind won't export it
	Password string
}

// GenerateConfig returns a configuration populated with platform defaults and
// a new identity.
func GenerateConfig() *NodeConfig {
	defaults := GetDefaults()
	cfg := new(NodeConfig)
	cfg.NewPrivateKey()
	cfg.Listen = []string{}
	cfg.AdminListen = defaults.DefaultAdminListen
	cfg.Peers = []string{}
	cfg.InterfacePeers = map[string][]string{}
	cfg.AllowedPublicKeys = []string{}
	cfg.MulticastInterfaces = defaults.DefaultMulticastInterfaces
	cfg.IfName = defaults.DefaultIfName
	cfg.IfMTU = defaults.DefaultIfMTU
	cfg.NodeInfoPrivacy = false
	if err := cfg.postprocessConfig(); err != nil {
		panic(err)
	}
	return cfg
}

// ReadFrom replaces cfg with configuration parsed from r. Missing settings use
// platform defaults, but persistent identity is always required.
func (cfg *NodeConfig) ReadFrom(r io.Reader) (int64, error) {
	conf, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	n := int64(len(conf))
	// Decode UTF-16 configuration with a byte order mark before HJSON parsing.
	if len(conf) >= 2 && (bytes.Equal(conf[0:2], []byte{0xFF, 0xFE}) ||
		bytes.Equal(conf[0:2], []byte{0xFE, 0xFF})) {
		utf := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
		decoder := utf.NewDecoder()
		conf, err = decoder.Bytes(conf)
		if err != nil {
			return n, err
		}
	}
	// Generate a new configuration - this gives us a set of sane defaults -
	// then parse the configuration we loaded above on top of it. The effect
	// of this is that any configuration item that is missing from the provided
	// configuration will use a sane default, except persistent identity.
	*cfg = *GenerateConfig()
	cfg.PrivateKey = nil
	if err := cfg.UnmarshalHJSON(conf); err != nil {
		return n, err
	}
	return n, nil
}

// UnmarshalHJSON parses HJSON or JSON into cfg and validates its identity.
func (cfg *NodeConfig) UnmarshalHJSON(b []byte) error {
	if err := hjson.Unmarshal(b, cfg); err != nil {
		return err
	}
	return cfg.postprocessConfig()
}

func (cfg *NodeConfig) postprocessConfig() error {
	if cfg.PrivateKeyPath != "" {
		cfg.PrivateKey = nil
		f, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return err
		}
		if err := cfg.UnmarshalPEMPrivateKey(f); err != nil {
			return err
		}
	}
	if len(cfg.PrivateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("configuration requires a valid PrivateKey or PrivateKeyPath; refusing to generate a replacement identity")
	}

	switch {
	case cfg.Certificate == nil:
		// No self-signed certificate has been generated yet.
		fallthrough
	case !bytes.Equal(cfg.Certificate.PrivateKey.(ed25519.PrivateKey), cfg.PrivateKey):
		// A self-signed certificate was generated but the private
		// key has changed since then, possibly because a new config
		// was parsed.
		if err := cfg.GenerateSelfSignedCertificate(); err != nil {
			return err
		}
	}
	return nil
}

// RFC5280 section 4.1.2.5
var notAfterNeverExpires = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)

// GenerateSelfSignedCertificate derives the transport certificate from cfg's
// persistent identity key.
func (cfg *NodeConfig) GenerateSelfSignedCertificate() error {
	key, err := cfg.MarshalPEMPrivateKey()
	if err != nil {
		return err
	}
	cert, err := cfg.MarshalPEMCertificate()
	if err != nil {
		return err
	}
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return err
	}
	cfg.Certificate = &tlsCert
	return nil
}

// MarshalPEMCertificate encodes cfg's self-signed certificate as PEM.
func (cfg *NodeConfig) MarshalPEMCertificate() ([]byte, error) {
	privateKey := ed25519.PrivateKey(cfg.PrivateKey)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: hex.EncodeToString(publicKey),
		},
		NotBefore:             time.Now(),
		NotAfter:              notAfterNeverExpires,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certbytes, err := x509.CreateCertificate(rand.Reader, cert, cert, publicKey, privateKey)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certbytes,
	}
	return pem.EncodeToMemory(block), nil
}

// NewPrivateKey replaces cfg's identity with a newly generated Ed25519 key.
func (cfg *NodeConfig) NewPrivateKey() {
	_, spriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	cfg.PrivateKey = KeyBytes(spriv)
}

// MarshalPEMPrivateKey encodes cfg's identity key as PKCS#8 PEM.
func (cfg *NodeConfig) MarshalPEMPrivateKey() ([]byte, error) {
	b, err := x509.MarshalPKCS8PrivateKey(ed25519.PrivateKey(cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PKCS8 key: %w", err)
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}
	return pem.EncodeToMemory(block), nil
}

// UnmarshalPEMPrivateKey loads an Ed25519 PKCS#8 PEM identity into cfg.
func (cfg *NodeConfig) UnmarshalPEMPrivateKey(b []byte) error {
	p, _ := pem.Decode(b)
	if p == nil {
		return fmt.Errorf("failed to parse PEM file")
	}
	if p.Type != "PRIVATE KEY" {
		return fmt.Errorf("unexpected PEM type %q", p.Type)
	}
	k, err := x509.ParsePKCS8PrivateKey(p.Bytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal PKCS8 key: %w", err)
	}
	key, ok := k.(ed25519.PrivateKey)
	if !ok {
		return fmt.Errorf("private key must be ed25519 key")
	}
	if len(key) != ed25519.PrivateKeySize {
		return fmt.Errorf("unexpected ed25519 private key length")
	}
	cfg.PrivateKey = KeyBytes(key)
	return nil
}

// KeyBytes marshals private-key bytes as hexadecimal JSON text.
type KeyBytes []byte

// MarshalJSON implements json.Marshaler.
func (k KeyBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(k))
}

// UnmarshalJSON implements json.Unmarshaler.
func (k *KeyBytes) UnmarshalJSON(b []byte) error {
	var s string
	var err error
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}
	*k, err = hex.DecodeString(s)
	return err
}
