package peerutils

import (
	"crypto/rsa"
	"errors"
	"net"

	"github.com/DrewRoss5/courier/cryptoutils"
)

type Tunnel struct {
	sessionKey []byte
	PeerPubKey rsa.PublicKey
	userPrvKey rsa.PrivateKey
	Incoming   net.Conn
	Outgoing   net.Conn
	Peer       User
	User       User
}

// encrypts and sends the provided message through this Tunnel
func (t Tunnel) SendMessage(message []byte) error {
	// encrypt the plaintext
	ciphertext, err := cryptoutils.AesEncrypt(message, t.sessionKey)
	if err != nil {
		return err
	}
	// create a signature of the plaintext message and append it to the ciphertext
	signature, err := cryptoutils.RsaSign(t.userPrvKey, message)
	if err != nil {
		return err
	}
	message = append(signature, ciphertext...)
	// send the message and get the response
	responseBuf := make([]byte, 1)
	t.Outgoing.Write(message)
	t.Outgoing.Read(responseBuf)
	if responseBuf[0] != 0x0 {
		return errors.New("message not validated")
	}
	return nil
}

func (t Tunnel) AwaitMessage() ([]byte, error) {
	_, messageRaw, err := RecvAll(t.Incoming)
	if err != nil {
		t.Incoming.Write([]byte{RES_ERR})
		return nil, err
	}
	signature := messageRaw[:cryptoutils.SIGNATURE_SIZE]
	cipherext := messageRaw[cryptoutils.SIGNATURE_SIZE:]
	message, err := cryptoutils.AesDecrypt(cipherext, t.sessionKey)
	if err != nil {
		t.Incoming.Write([]byte{RES_ERR})
		return nil, err
	}
	message = cryptoutils.StripZeroes(message)
	if !cryptoutils.RsaVerify(t.PeerPubKey, message, signature) {
		t.Incoming.Write([]byte{RES_ERR})
		return nil, errors.New("failed to verify RSA signature")
	}
	// send the response to the peer
	t.Incoming.Write([]byte{RES_OK})
	return message, nil
}

// quits the connection and sends a message to the peer indicating such
func (t Tunnel) Shutdown() error {
	// send a message to the peer indicating that this Tunnel is being closed
	err := t.SendMessage([]byte{MESSAGE_DISCONNECT})
	if err != nil {
		return err
	}
	t.Incoming.Close()
	t.Outgoing.Close()
	return nil
}
