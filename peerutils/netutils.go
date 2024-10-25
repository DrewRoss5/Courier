package peerutils

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/DrewRoss5/courier/cryptoutils"
)

const BUF_SIZE = 1024

// message codes wil be defined here
const (
	RES_OK             byte = 0x0
	RES_ERR            byte = 0x1
	MESSAGE_INIT       byte = 0x1
	MESSAGE_TXT        byte = 0x2
	MESSAGE_DISCONNECT byte = 0x3
	CHAT_ARCHIVE       byte = 0x4
)

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
		if err != nil {
			return 0, nil, err
		}
		totalRead += sizeRead
		message = append(message, buf[:sizeRead]...)
		if sizeRead < BUF_SIZE {
			break
		}
	}
	return totalRead, message, nil
}

// attempts to connect to a peer, returning a tunnel if the peer can be reached
func ConnectPeer(addr string, pubKey rsa.PublicKey, prvKey rsa.PrivateKey, initiator User) (*Tunnel, error) {
	conn, err := net.Dial("tcp", addr+":54000")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// send this RSA key, and await the response
	conn.Write(append([]byte{MESSAGE_INIT}, cryptoutils.ExportRsaPub(&pubKey)...))
	_, response, err := RecvAll(conn)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	// peer does not initiate the connection
	if response[0] != RES_OK {
		return nil, errors.New("connection not established")
	}
	// parse the public key
	peerPub, err := cryptoutils.ImportRsaPub(response[1:])
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	// generate, encrypt, and send the session key
	sessionKey := cryptoutils.GenAesKey()
	keyCiphertext, _ := cryptoutils.RsaEncrypt(&peerPub, sessionKey)
	conn.Write(keyCiphertext)
	conn.Read(response)
	if response[0] != RES_OK {
		conn.Write([]byte{RES_ERR})
		return nil, errors.New("failed to verify the session key with peer")
	}
	// create and encrypt a challegene
	checksum := cryptoutils.GenNonce()
	checksumCiphertext, _ := cryptoutils.AesEncrypt(checksum, sessionKey)
	_, err = conn.Write(checksumCiphertext)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	_, checksumResponse, err := RecvAll(conn)
	if err != nil {
		return nil, err
	}
	responsePlaintext, err := cryptoutils.RsaDecrypt(&prvKey, checksumResponse)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	if !compSlices(checksum, responsePlaintext) {
		conn.Write([]byte{RES_ERR})
		return nil, errors.New("failed to verify the session key with peer")
	}
	// send the information of this user to the peer
	userJson, err := json.Marshal(initiator)
	if err != nil {
		return nil, err
	}
	userCiphertext, _ := cryptoutils.AesEncrypt(userJson, sessionKey)
	_, err = conn.Write(append([]byte{RES_OK}, userCiphertext...))
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
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	var peer User
	err = json.Unmarshal(cryptoutils.StripZeroes(peerInfo), &peer)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	// validate the peer's ID
	if !ValidateId(peer.Id, &peerPub) {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	// validate the peer's username and color
	if !slices.Contains([]string{Red, Green, Blue, Yellow, Magenta, Cyan, Gray, White}, peer.Color) {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	if len(peer.Name) > 64 {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	_, err = conn.Write([]byte{RES_OK})
	if err != nil {
		return nil, err
	}
	// await a connection on this incoming port
	listener, err := net.Listen("tcp", addr+":54001")
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	incoming, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	// connect to the peer's incoming port
	time.Sleep(time.Second) // this is a very hacky way of avoiding a race condition and needs to be fixed
	outgoing, err := net.Dial("tcp", addr+":54002")
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	return &Tunnel{sessionKey: sessionKey, PeerPubKey: peerPub, userPrvKey: prvKey, Incoming: incoming, Outgoing: outgoing, Peer: peer, User: initiator}, nil
}

func AwaitPeer(pubKey rsa.PublicKey, prvKey rsa.PrivateKey, reciever User) (*Tunnel, error) {
	listener, err := net.Listen("tcp", ":54000")
	if err != nil {
		return nil, err
	}
	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, message, err := RecvAll(conn)
	if err != nil {
		return nil, err
	}
	// TODO: implement additional features for differing requests
	if message[0] != MESSAGE_INIT {
		return nil, errors.New("unrecognized request")
	}
	peerPem := message[1:]
	peerPub, err := cryptoutils.ImportRsaPub(peerPem)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	conn.Write(append([]byte{RES_OK}, cryptoutils.ExportRsaPub(&pubKey)...))
	// await the session key
	_, keyCiphertext, err := RecvAll(conn)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	sessionKey, err := cryptoutils.RsaDecrypt(&prvKey, keyCiphertext)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	conn.Write([]byte{RES_OK})
	// await the challenge
	_, challenge, err := RecvAll(conn)
	if err != nil {
		return nil, err
	}
	challenge, err = cryptoutils.AesDecrypt(challenge, sessionKey)
	challenge = challenge[len(challenge)-16:]
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	challengeResponse, err := cryptoutils.RsaEncrypt(&peerPub, challenge)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	conn.Write(challengeResponse)
	// await the verification
	_, response, err := RecvAll(conn)
	if err != nil {
		return nil, err
	}
	if response[0] != RES_OK {
		return nil, errors.New("failed to verify the session with peer ")
	}
	response = response[1:]
	// recieve the peer's information
	peerInfo, err := cryptoutils.AesDecrypt(response, sessionKey)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	var peer User
	err = json.Unmarshal(cryptoutils.StripZeroes(peerInfo), &peer)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	// validate the peer's ID
	if !ValidateId(peer.Id, &peerPub) {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	// send the peer the user's information
	userInfo, _ := json.Marshal(reciever)
	userCiphertext, _ := cryptoutils.AesEncrypt(userInfo, sessionKey)
	conn.Write(append([]byte{RES_OK}, userCiphertext...))
	_, response, err = RecvAll(conn)
	if err != nil {
		conn.Write([]byte{RES_ERR})
		return nil, err
	}
	if response[0] != RES_OK {
		return nil, errors.New("failed to initiate the connection")
	}
	// connect to the peer's incoming port
	peerAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
	time.Sleep(time.Second) // this is a very hacky way of avoiding a race condition and needs to be fixed
	outgoing, err := net.Dial("tcp", peerAddr+":54001")
	if err != nil {
		return nil, err
	}
	// await the peer's connection to the incoming port
	listener, err = net.Listen("tcp", peerAddr+":54002")
	if err != nil {
		return nil, err
	}
	incoming, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	return &Tunnel{sessionKey: sessionKey, PeerPubKey: peerPub, userPrvKey: prvKey, Incoming: incoming, Outgoing: outgoing, Peer: peer, User: reciever}, nil
}
