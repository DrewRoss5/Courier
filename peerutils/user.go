package peerutils

import (
	"crypto/rsa"
	"encoding/base64"

	"github.com/DrewRoss5/courier/cryptoutils"
)

const (
	ID_SIZE = 512
)

type User struct {
	Name  string
	Color string
	Id    string
}

// given a user's private key, generates their id, in the form of a Base64-encoded signature of a constant
func GenId(prvKey *rsa.PrivateKey) (string, error) {
	sigConst := []byte{108, 14, 41, 125, 134, 205, 20, 243, 160, 29, 4, 162, 61, 70, 126, 246} // this constant isn't important, and just serves an arbitrary value to generate a signature of
	idBytes, err := cryptoutils.RsaSign(*prvKey, sigConst)
	if err != nil {
		return "", err
	}
	idStr := base64.StdEncoding.EncodeToString(idBytes)
	return idStr, nil

}

// verifies a peer's ID given their public key
func ValidateId(id string, pubKey *rsa.PublicKey) bool {
	// decode the base64-encoded signature
	idBytes := make([]byte, ID_SIZE)
	_, err := base64.StdEncoding.Decode(idBytes, []byte(id))
	if err != nil {
		return false
	}
	// validate the signature
	sigConst := []byte{108, 14, 41, 125, 134, 205, 20, 243, 160, 29, 4, 162, 61, 70, 126, 246}
	return cryptoutils.RsaVerify(*pubKey, sigConst, idBytes)
}
