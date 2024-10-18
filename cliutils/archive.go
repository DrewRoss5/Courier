package cliutils

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/DrewRoss5/courier/cryptoutils"
	"github.com/DrewRoss5/courier/peerutils"
)

const (
	ROUNDS    = 16
	SALT_SIZE = 12
)

// generates an AES key from a password, by hashing it with a salt repeatedly
func hashKey(password []byte, salt []byte) []byte {
	key := password
	for i := 0; i < ROUNDS; i++ {
		key = cryptoutils.HashKey(key, salt)
	}
	return key
}

// archives a chat and saves it to a file in a given path
func ArchiveChat(chat peerutils.Chatroom, password string, path string) error {
	// create the path if needed
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}
	// generate a key
	salt := cryptoutils.GenNonce()
	key := hashKey([]byte(password), salt)
	// write the chat to a buffer, and encrypt it
	var buf bytes.Buffer
	chat.DisplayMessages(&buf)
	ciphertext, err := cryptoutils.AesEncrypt(buf.Bytes(), key)
	if err != nil {
		return err
	}
	// save the salt and ciphertext to the file
	contents := append(salt, ciphertext...)
	fileName := time.Now().Format("15-04-05.arc")
	file, err := os.Open(path + fileName)
	if err != nil {
		return err
	}
	file.Write(contents)
	return nil
}

// returns a string of a decrypted chat archive
func DecryptArchive(fileName string, password string) (string, error) {
	// read and parse the file
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	if len(fileContents) < cryptoutils.AES_MIN_CIPHERTEXT_SIZE+SALT_SIZE {
		return "", errors.New("archive: invalid archive file")
	}
	salt := fileContents[:SALT_SIZE]
	ciphertext := fileContents[SALT_SIZE:]
	// generate the key and attempt to decrypt the archive
	key := hashKey([]byte(password), salt)
	archive, err := cryptoutils.AesDecrypt(ciphertext, key)
	if err != nil {
		return "", err
	}
	return string(archive), nil
}
