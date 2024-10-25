package cryptoutils

import "crypto/sha256"

// generates an AES key from a password, by hashing it with a salt repeatedly
func HashKey(password []byte, salt []byte, rounds int) []byte {
	key := password
	for i := 0; i < rounds; i++ {
		hasher := sha256.New()
		hasher.Write(key)
		hasher.Write(salt)
		key = hasher.Sum(nil)
	}
	return key
}
