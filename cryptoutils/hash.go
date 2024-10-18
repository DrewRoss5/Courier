package cryptoutils

import "crypto/sha256"

// creates a sha256 to be used as a key, given plaintext and salt
func HashKey(plaintext []byte, salt []byte) []byte {
	hasher := sha256.New()
	hasher.Write(plaintext)
	hasher.Write(salt)
	return hasher.Sum(nil)
}
