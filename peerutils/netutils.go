package peerutils

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net"

	"github.com/DrewRoss5/courier/cryptoutils"
)

const BUF_SIZE = 1024

// message codes wil be defined here
const RES_OK byte = 0x0
const MESSAGE_INIT byte = 0x1
const MESSAGE_TXT byte = 0x2
const DISCONNECT byte = 0x3

// a helper function to compare two byte slices
func compSlices(a []byte, b []byte) bool {
	sliceLen := len(a)
	if sliceLen != len(b) {
		return false
	}
	for i := 0; i < sliceLen; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// recieves data of unknown size from Conn object
func RecvAll(conn net.Conn) (int, []byte, error) {
	message := make([]byte, 0, BUF_SIZE)
	buf := make([]byte, BUF_SIZE)
	totalRead := 0
	for {
		sizeRead, err := conn.Read(buf)
		if sizeRead == 0 {
			break
		}
		if err != nil {
			return 0, nil, err
		}
		totalRead += sizeRead
		message = append(message, buf[:sizeRead]...)
	}
	return totalRead, message, nil
}

// attempts to connect to a peer, returning a tunnel if the peer can be reached
func ConnectPeer(addr string, pubKey rsa.PublicKey, prvKey rsa.PrivateKey, initiator user) (*tunnel, error) {
	conn, err := net.Dial("tcp", addr+":54000")
	if err != nil {
		return nil, err
	}
	// send this RSA key, and await the response
	conn.Write(append([]byte{MESSAGE_INIT}, cryptoutils.ExportRsaPub(&pubKey)...))
	responseSize, response, err := RecvAll(conn)
	if err != nil {
		return nil, err
	}
	// peer does not initiate the connection
	if response[0] != RES_OK {
		return nil, errors.New("connection not established")
	}
	if responseSize < cryptoutils.RSA_KEY_SIZE {
		return nil, errors.New("recieved invalid RSA public key")
	}
	// parse the public key
	peerPub, err := cryptoutils.ImportRsaPub(response[1:])
	if err != nil {
		return nil, err
	}
	// generate, encrypt, and send the session key
	sessionKey := cryptoutils.GenAesKey()
	keyCiphertext, _ := cryptoutils.RsaEncrypt(&peerPub, sessionKey)
	conn.Write(keyCiphertext)
	conn.Read(response)
	if response[0] != RES_OK {
		return nil, errors.New("failed to verify the session key with peer")
	}
	// create and encrypt a challegene
	checksum := cryptoutils.GenNonce()
	checksumCiphertext, _ := cryptoutils.AesEncrypt(checksum, sessionKey)
	checksumResponse := make([]byte, 16)
	_, err = conn.Write(checksumCiphertext)
	if err != nil {
		return nil, err
	}
	_, err = conn.Read(checksumResponse)
	if err != nil {
		return nil, err
	}
	responsePlaintext, err := cryptoutils.AesDecrypt(checksumResponse, sessionKey)
	if err != nil {
		return nil, err
	}
	if !compSlices(checksum, responsePlaintext) {
		return nil, errors.New("failed to verify the session key with peer")
	}
	// send the information of this user to the peer
	userJson, err := json.Marshal(initiator)
	if err != nil {
		return nil, err
	}
	userCiphertext, _ := cryptoutils.AesEncrypt(userJson, sessionKey)
	_, err = conn.Write(userCiphertext)
	if err != nil {
		return nil, err
	}
	// recieve and decrypt the peer's info
	_, peerCiphertext, err := RecvAll(conn)
	if err != nil || peerCiphertext[0] != RES_OK {
		return nil, err
	}
	peerInfo, err := cryptoutils.AesDecrypt(peerCiphertext[1:], sessionKey)
	if err != nil {
		return nil, err
	}
	var peer user
	err = json.Unmarshal(peerInfo, &peer)
	if err != nil {
		return nil, err
	}
	response = make([]byte, 1)
	response[0] = RES_OK
	_, err = conn.Write(response)
	if err != nil {
		return nil, err
	}
	// await a connection on this incoming port
	listener, err := net.Listen("tcp", addr+":54001")
	if err != nil {
		return nil, err
	}
	incoming, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	// connect to the peer's incoming port
	outgoing, err := net.Dial("tcp", addr+":54002")
	if err != nil {
		return nil, err
	}
	return &tunnel{sessionKey: sessionKey, peerPubKey: peerPub, userPrvKey: prvKey, incoming: incoming, outgoing: outgoing, peer: peer, user: initiator}, nil
}
