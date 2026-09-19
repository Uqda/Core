/*
This file generates crypto keys for [ansible-yggdrasil](https://github.com/jcgruenhage/ansible-yggdrasil/)
*/
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/Uqda/Core/src/address"
	"github.com/cheggaaa/pb/v3"
)

var numHosts = flag.Int("hosts", 1, "number of host vars to generate")
var keyTries = flag.Int("tries", 1000, "number of tries before taking the best keys")

type keySet struct {
	priv []byte
	pub  []byte
	ip   string
}

func main() {
	flag.Parse()

	bar := pb.StartNew(*keyTries*2 + *numHosts)

	if *numHosts > *keyTries {
		println("Can't generate less keys than hosts.")
		return
	}

	var keys []keySet
	for i := 0; i < *numHosts+1; i++ {
		keys = append(keys, newKey())
		bar.Increment()
	}
	keys = sortKeySetArray(keys)
	for i := 0; i < *keyTries-*numHosts-1; i++ {
		keys[0] = newKey()
		keys = bubbleUpTo(keys, 0)
		bar.Increment()
	}

	_ = os.MkdirAll("host_vars", 0755)

	for i := 1; i <= *numHosts; i++ {
		if err := writeHostVars(fmt.Sprintf("host_vars/%x", i), keys[i]); err != nil {
			fmt.Println("Error writing host_vars:", err)
			return
		}
		bar.Increment()
	}
	bar.Finish()
}

// writeHostVars writes the ansible vars file (public key and IP - not
// sensitive) and the vault file (the raw private key) for one host. The
// vault file is created with 0600 permissions and explicitly chmod'd to
// 0600 afterwards, since os.WriteFile's mode argument only applies when it
// creates a brand new file - re-running this tool against an existing
// world-readable vault file would otherwise leave its permissions
// unchanged.
func writeHostVars(dir string, key keySet) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	vars := fmt.Sprintf(
		"yggdrasil_public_key: %v\nyggdrasil_private_key: \"{{ vault_yggdrasil_private_key }}\"\nansible_host: %v\n",
		hex.EncodeToString(key.pub), key.ip)
	if err := os.WriteFile(filepath.Join(dir, "vars"), []byte(vars), 0644); err != nil {
		return err
	}

	vault := fmt.Sprintf("vault_yggdrasil_private_key: %v\n", hex.EncodeToString(key.priv))
	vaultPath := filepath.Join(dir, "vault")
	if err := os.WriteFile(vaultPath, []byte(vault), 0600); err != nil {
		return err
	}
	return os.Chmod(vaultPath, 0600)
}

func newKey() keySet {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	ip := net.IP(address.AddrForKey(pub)[:]).String()
	return keySet{priv[:], pub[:], ip}
}

func isBetter(oldID, newID []byte) bool {
	for idx := range oldID {
		if newID[idx] < oldID[idx] {
			return true
		}
		if newID[idx] > oldID[idx] {
			return false
		}
	}
	return false
}

func sortKeySetArray(sets []keySet) []keySet {
	for i := 0; i < len(sets); i++ {
		sets = bubbleUpTo(sets, i)
	}
	return sets
}

func bubbleUpTo(sets []keySet, num int) []keySet {
	for i := 0; i < len(sets)-num-1; i++ {
		if isBetter(sets[i+1].pub, sets[i].pub) {
			var tmp = sets[i]
			sets[i] = sets[i+1]
			sets[i+1] = tmp
		} else {
			break
		}
	}
	return sets
}
