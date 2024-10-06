package cryptoutils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const RSA_KEY_SIZE = 4096

// generates a pair of RSA keys and saves them to the provided path
func GenerateRsaKeys(keyPath string) error {
	// generate the rsa keys
	priv, _ := rsa.GenerateKey(rand.Reader, RSA_KEY_SIZE)
	//save the private key
	prvBytes := x509.MarshalPKCS1PrivateKey(priv)
	prvBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: prvBytes,
	}
	prvFile, err := os.Create(fmt.Sprintf("%v/prv.pem", keyPath))
	if err != nil {
		return err
	}
	defer prvFile.Close()
	_, err = prvFile.Write(pem.EncodeToMemory(&prvBlock))
	if err != nil {
		return err
	}
	// save the public key
	pubBytes := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	pubBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	}
	pubFile, err := os.Create(fmt.Sprintf("%v/pub.pem", keyPath))
	if err != nil {
		return err
	}
	defer pubFile.Close()
	_, err = pubFile.Write(pem.EncodeToMemory(&pubBlock))
	return err
}

// imports a pair of RSA keys from a file path
func ImportRsa(keyPath string) (rsa.PrivateKey, rsa.PublicKey, error) {
	// read the private key
	prvFile, err := os.Open(fmt.Sprintf("%v/prv.pem", keyPath))
	if err != nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, err
	}
	buf := make([]byte, RSA_KEY_SIZE)
	_, err = prvFile.Read(buf)
	if err != nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, err
	}
	prvBlock, _ := pem.Decode(buf)
	if prvBlock == nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, errors.New("invalid RSA private key")
	}
	prvKey, err := x509.ParsePKCS1PrivateKey(prvBlock.Bytes)
	if err != nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, err
	}
	// read the public key
	pubFile, err := os.Open(fmt.Sprintf("%v/pub.pem", keyPath))
	if err != nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, err
	}
	pubBuf := make([]byte, RSA_KEY_SIZE)
	_, err = pubFile.Read(pubBuf)
	if err != nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, err
	}
	pubBlock, _ := pem.Decode(pubBuf)
	if pubBlock == nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, errors.New("invalid RSA private key")
	}
	pubKey, err := x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	if err != nil {
		return rsa.PrivateKey{}, rsa.PublicKey{}, err
	}
	return *prvKey, *pubKey, nil
}

func ImportRsaPub(pemStr []byte) (rsa.PublicKey, error) {
	pubBlock, _ := pem.Decode(pemStr)
	if pubBlock == nil {
		return rsa.PublicKey{}, errors.New("invalid RSA public key")
	}
	pubKey, err := x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	if err != nil {
		return rsa.PublicKey{}, err
	}
	return *pubKey, err
}

func RsaEncrypt(pubKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plaintext)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

func RsaDecrypt(prvKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	plaintext, err := rsa.DecryptPKCS1v15(nil, prvKey, ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func RsaSign(prvKey rsa.PrivateKey, plaintext []byte) ([]byte, error) {
	hasher := crypto.SHA256.New()
	hasher.Write(plaintext)
	msg_hash := hasher.Sum(nil)
	signature, err := rsa.SignPKCS1v15(nil, &prvKey, crypto.SHA256, msg_hash)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func RsaVerify(pubKey rsa.PublicKey, plaintext []byte, signature []byte) bool {
	hasher := crypto.SHA256.New()
	hasher.Write(plaintext)
	msg_hash := hasher.Sum(nil)
	err := rsa.VerifyPKCS1v15(&pubKey, crypto.SHA256, msg_hash, signature)
	return err == nil
}
