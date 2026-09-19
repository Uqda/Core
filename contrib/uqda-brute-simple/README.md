# uqda-brute-simple

Simple program for finding curve25519 and ed25519 public keys whose sha512 hash has many leading ones.
Because ed25519 private keys consist of a seed that is hashed to find the secret part of the keypair,
this program is near optimal for finding ed25519 keypairs. Curve25519 key generation, on the other hand,
could be further optimized with elliptic curve magic.

Depends on libsodium.

This utility is inherited from upstream Yggdrasil (`yggdrasil-brute-simple`).
Uqda uses local product names for its directory, sources and executable targets;
upstream authorship and the original [LICENSE](LICENSE) are preserved.
The curve25519 variant is a historical tool, not the current Uqda identity format.
