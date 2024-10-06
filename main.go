package main

import (
	"encoding/base64"
	"fmt"

	"github.com/DrewRoss5/courier/cryptoutils"
)

func main() {
	fmt.Println("--------- AES TESTS ---------")
	key := cryptoutils.GenAesKey()
	plaintext := []byte("Real big secrets")
	ciphertext, _ := cryptoutils.AesEncrypt(plaintext, key)
	fmt.Printf("Key: %v\nEncrypted: %v\n", base64.StdEncoding.EncodeToString(key), base64.StdEncoding.EncodeToString(ciphertext))
	plaintext, _ = cryptoutils.AesDecrypt(ciphertext, key)
	fmt.Printf("Plaintext: %v\n", string(plaintext))
	fmt.Println("--------- RSA TESTS ---------")
	fmt.Println("Generating keys...")
	cryptoutils.GenerateRsaKeys("keys")
	fmt.Println("Generated")
	fmt.Println("Importing keys...")
	prvKey, pubKey, err := cryptoutils.ImportRsa("keys")
	if err != nil {
		fmt.Printf("Cannot read the keys. %v", err.Error())
		return
	}
	fmt.Println("Keys read")
	// test RSA encryption
	ciphertext, _ = cryptoutils.RsaEncrypt(&pubKey, []byte("Rsa shit"))
	plaintext, _ = cryptoutils.RsaDecrypt(&prvKey, ciphertext)
	fmt.Printf("Ciphertext: %v\nPlaintext: %v\n", base64.StdEncoding.EncodeToString(ciphertext), string(plaintext))
	// test rsa signing
	message := []byte("The british are coming!")
	signature, err := cryptoutils.RsaSign(prvKey, message)
	if err != nil {
		fmt.Printf("Signature error: %v\n", err.Error())
	}
	valid := cryptoutils.RsaVerify(pubKey, message, signature)
	if !valid {
		fmt.Println("Signature could not be validated")
		return
	}
	fmt.Printf("Signature: %v\n", base64.StdEncoding.EncodeToString(signature))

}
