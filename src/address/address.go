// Package address contains the types used by Uqda to represent IPv6 addresses or prefixes, as well as functions for working with these types.
// Of particular importance are the functions used to derive addresses or subnets from a NodeID, or to get the NodeID and bitmask of the bits visible from an address, which is needed for DHT searches.
package address

import (
	"crypto/ed25519"
)

const (
	prefixLength          = 1
	maxAddressLeadingOnes = 255
)

// Address represents an IPv6 address in the Yggdrasil network address range.
type Address [16]byte

// Subnet represents an IPv6 /64 subnet in the Yggdrasil network subnet range.
type Subnet [8]byte

// GetPrefix returns the address prefix used by the Yggdrasil network.
// The current implementation requires this to be a multiple of 8 bits + 7 bits.
// The 8th bit of the last byte is used to signal nodes (0) or /64 prefixes (1).
// Nodes that configure this differently will be unable to communicate with each other using IP packets, though routing and the DHT machinery *should* still work.
func GetPrefix() [prefixLength]byte {
	return [...]byte{0x02}
}

// IsValid returns true if an address falls within the range used by nodes in the network.
func (a *Address) IsValid() bool {
	prefix := GetPrefix()
	for idx := range prefix {
		if (*a)[idx] != prefix[idx] {
			return false
		}
	}
	return true
}

// IsValid returns true if a prefix falls within the range usable by the network.
func (s *Subnet) IsValid() bool {
	prefix := GetPrefix()
	l := len(prefix)
	for idx := range prefix[:l-1] {
		if (*s)[idx] != prefix[idx] {
			return false
		}
	}
	return (*s)[l-1] == prefix[l-1]|0x01
}

// AddrForKey derives an Address from an Ed25519 public key. It returns nil if
// the key length or its leading-bit encoding cannot be represented.
// This address begins with the contents of GetPrefix(), with the last bit set to 0 to indicate an address.
// The following 8 bits are set to the number of leading 1 bits in the bitwise inverse of the public key.
// The bitwise inverse of the key, excluding the leading 1 bits and the first leading 0 bit, is truncated to the appropriate length and makes up the remainder of the address.
func AddrForKey(publicKey ed25519.PublicKey) *Address {
	if len(publicKey) != ed25519.PublicKeySize {
		return nil
	}
	var buf [ed25519.PublicKeySize]byte
	copy(buf[:], publicKey)
	for idx := range buf {
		buf[idx] = ^buf[idx]
	}
	var addr Address
	temp := make([]byte, 0, 32)
	done := false
	ones := 0
	bits := byte(0)
	nBits := 0
	for idx := 0; idx < 8*len(buf); idx++ {
		bit := (buf[idx/8] & (0x80 >> byte(idx%8))) >> byte(7-(idx%8))
		if !done && bit != 0 {
			if ones == maxAddressLeadingOnes {
				return nil
			}
			ones++
			continue
		}
		if !done && bit == 0 {
			done = true
			continue
		}
		bits = (bits << 1) | bit
		nBits++
		if nBits == 8 {
			nBits = 0
			temp = append(temp, bits)
		}
	}
	prefix := GetPrefix()
	copy(addr[:], prefix[:])
	addr[len(prefix)] = byte(ones)
	copy(addr[len(prefix)+1:], temp)
	return &addr
}

// SubnetForKey takes an ed25519.PublicKey as an argument and returns a *Subnet.
// This function returns nil if the key length is not ed25519.PublicKeySize.
// The subnet begins with the address prefix, with the last bit set to 1 to indicate a prefix.
// The following 8 bits are set to the number of leading 1 bits in the bitwise inverse of the key.
// The bitwise inverse of the key, excluding the leading 1 bits and the first leading 0 bit, is truncated to the appropriate length and makes up the remainder of the subnet.
func SubnetForKey(publicKey ed25519.PublicKey) *Subnet {
	// Exactly as the address version, with two exceptions:
	//  1) The first bit after the fixed prefix is a 1 instead of a 0
	//  2) It's truncated to a subnet prefix length instead of 128 bits
	addr := AddrForKey(publicKey)
	if addr == nil {
		return nil
	}
	var snet Subnet
	copy(snet[:], addr[:])
	snet[prefixLength-1] |= 0x01
	return &snet
}

// GetKey returns the partial ed25519.PublicKey for the Address.
// This is used for key lookup.
func (a *Address) GetKey() ed25519.PublicKey {
	var key [ed25519.PublicKeySize]byte
	ones := int(a[prefixLength])
	for idx := 0; idx < ones; idx++ {
		key[idx/8] |= 0x80 >> byte(idx%8)
	}
	keyOffset := ones + 1
	addrOffset := 8*prefixLength + 8
	for idx := addrOffset; idx < 8*len(a); idx++ {
		bits := a[idx/8] & (0x80 >> byte(idx%8))
		bits <<= byte(idx % 8)
		keyIdx := keyOffset + (idx - addrOffset)
		bits >>= byte(keyIdx % 8)
		idx := keyIdx / 8
		if idx >= len(key) {
			break
		}
		key[idx] |= bits
	}
	for idx := range key {
		key[idx] = ^key[idx]
	}
	return ed25519.PublicKey(key[:])
}

// GetKey returns the partial ed25519.PublicKey for the Subnet.
// This is used for key lookup.
func (s *Subnet) GetKey() ed25519.PublicKey {
	var addr Address
	copy(addr[:], s[:])
	return addr.GetKey()
}
