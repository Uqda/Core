package config

import (
	"crypto/ed25519"
	"fmt"
	"runtime"
	"sync"
)

// MaxKeyStrength caps NewPrivateKeyWithStrength. Each step doubles the
// expected work; 24 means about 16 million keys, under a minute on a laptop.
const MaxKeyStrength = 24

// DefaultKeyStrength is used for identities created by uqda setup and join.
// It costs about 256 attempts, a few milliseconds.
const DefaultKeyStrength = 8

// KeyStrength is the number of leading zero bits in a public key. A Uqda
// address stores the key bits after them, so each extra bit puts one more
// bit of the key into the address and the /64 subnet. That doubles the work
// needed to generate another key with a colliding address or subnet.
func KeyStrength(pub ed25519.PublicKey) int {
	n := 0
	for _, b := range pub {
		if b == 0 {
			n += 8
			continue
		}
		for mask := byte(0x80); mask != 0 && b&mask == 0; mask >>= 1 {
			n++
		}
		break
	}
	return n
}

// NewPrivateKeyWithStrength replaces cfg's identity with a new key of at
// least the given strength, searching on every CPU.
func (cfg *NodeConfig) NewPrivateKeyWithStrength(strength int) error {
	if strength < 0 || strength > MaxKeyStrength {
		return fmt.Errorf("strength must be between 0 and %d", MaxKeyStrength)
	}
	found := make(chan ed25519.PrivateKey, 1)
	done := make(chan struct{})
	var once sync.Once
	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				pub, priv, err := ed25519.GenerateKey(nil)
				if err != nil {
					continue
				}
				if KeyStrength(pub) >= strength {
					once.Do(func() {
						found <- priv
						close(done)
					})
					return
				}
			}
		}()
	}
	priv := <-found
	wg.Wait()
	cfg.PrivateKey = KeyBytes(priv)
	cfg.PrivateKeyPath = ""
	return nil
}
