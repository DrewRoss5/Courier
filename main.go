package main

import (
	"fmt"
	"os"

	"github.com/DrewRoss5/courier/cryptoutils"
	"github.com/DrewRoss5/courier/peerutils"
)

func main() {
	// ensure the correct number of arguments
	if len(os.Args) != 3 {
		fmt.Println("This program accepts exactly two arguments ")
		return
	}
	username := os.Args[1]
	command := os.Args[2]
	// determine if the user is testing initating or recieving code
	switch command {
	case "create":
		fmt.Print("Key path: ")
		var keyPath string
		fmt.Scanf("%s", &keyPath)
		fmt.Println("Generating keys...")
		err := cryptoutils.GenerateRsaKeys(keyPath)
		if err != nil {
			fmt.Println("Error: could not create the keys. Ensure the selected path exists")
			return
		}
		fmt.Println("Keys created successfully")
	case "recieve":
		fmt.Print("Key path: ")
		var keyPath string
		fmt.Scanf("%s", &keyPath)
		prvKey, pubKey, err := cryptoutils.ImportRsa(keyPath)
		if err != nil {
			fmt.Println("Error: failed to  import the rsa keys")
			return
		}
		user := peerutils.User{Name: username, Color: "white"}
		fmt.Println("Awaiting connection...")
		tunnel, err := peerutils.AwaitPeer(pubKey, prvKey, user)
		if err != nil {
			fmt.Printf("Error: %v\n", err.Error())
			return
		}
		fmt.Printf("Established connection with %v at %v\n", tunnel.Peer.Name, tunnel.Incoming.RemoteAddr().String())
	case "initiate":
		fmt.Print("Key path: ")
		var keyPath string
		fmt.Scanf("%s", &keyPath)
		prvKey, pubKey, err := cryptoutils.ImportRsa(keyPath)
		if err != nil {
			fmt.Println("Error: failed to  import the rsa keys")
			return
		}
		var addr string
		fmt.Print("Peer address: ")
		fmt.Scanf("%s", addr)
		user := peerutils.User{Name: username, Color: "white"}
		tunnel, err := peerutils.ConnectPeer(addr, pubKey, prvKey, user)
		if err != nil {
			fmt.Printf("Error: %v\n", err.Error())
			return
		}
		fmt.Printf("Established connection with %v at %v\n", tunnel.Peer.Name, tunnel.Incoming.RemoteAddr().String())

	default:
		fmt.Println("Unrecognized command")
	}

}
