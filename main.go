package main

import (
	"encoding/base64"
	"fmt"

	"github.com/DrewRoss5/courier/cryptoutils"
)

func main() {
	key := cryptoutils.GenAesKey()
	plaintext := []byte("Real big secrets")
	ciphertext, _ := cryptoutils.AesEncrypt(plaintext, key)
	fmt.Printf("Key: %v\nEncrypted: %v\n", base64.StdEncoding.EncodeToString(key), base64.StdEncoding.EncodeToString(ciphertext))
	plaintext, _ = cryptoutils.AesDecrypt(ciphertext, key)
	fmt.Printf("Plaintext: %v\n", string(plaintext))
}
