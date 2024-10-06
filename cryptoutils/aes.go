package cryptoutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

const (
	AES_KEY_SIZE            = 32
	AES_MIN_CIPHERTEXT_SIZE = aes.BlockSize + 12
)

// returns a radnomly generated 256-bit aes key
func GenAesKey() []byte {
	key := make([]byte, AES_KEY_SIZE)
	rand.Reader.Read(key)
	return key
}

// encrypts a provided plaintext with AES256 in GCM mode, and appends a nonce to the start of the ciphertext
func AesEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	// validate the size of the key
	if len(key) != AES_KEY_SIZE {
		return nil, errors.New("invalid AES key size")
	}
	// create a cipher
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, _ := cipher.NewGCM(aesCipher)
	// generate a nonce
	nonce := make([]byte, gcm.NonceSize())
	rand.Reader.Read(nonce)
	ciphertext := gcm.Seal(plaintext[:0], nonce, plaintext, nil)
	ciphertext = append(nonce, ciphertext...)
	return ciphertext, nil
}

// decrypts a ciphertext encrypted with AesEncrypt
func AesDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// validate the size of the key and ciphertext
	if len(key) != AES_KEY_SIZE {
		return nil, errors.New("invalid AES key size")
	}
	if len(ciphertext) < AES_MIN_CIPHERTEXT_SIZE || ((len(ciphertext)-12)%aes.BlockSize) != 0 {
		return nil, errors.New("invalid AES ciphertext size")
	}
	// create the cipher
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, _ := cipher.NewGCM(aesCipher)
	// parse the ciphertext into a nonce and the rest of the ciphertext
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	// attempt to decrypt the plaintext
	plaintext := make([]byte, len(ciphertext)-gcm.NonceSize())
	plaintext, err = gcm.Open(plaintext, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
