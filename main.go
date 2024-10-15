package main

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strings"

	"github.com/DrewRoss5/courier/cryptoutils"

	"github.com/DrewRoss5/courier/cliutils"
	"github.com/DrewRoss5/courier/peerutils"
)

// there aren't real accounts, but this creates a user for "login"
func login() (rsa.PrivateKey, rsa.PublicKey, peerutils.User) {
	// read the user's key
	fmt.Print("Key path: ")
	var keyPath string
	fmt.Scanf("%s", &keyPath)
	fmt.Println("Importing RSA keys...")
	prvKey, pubKey, err := cryptoutils.ImportRsa(keyPath)
	if err != nil {
		fmt.Println("Failed to import RSA keys. Exiting...")
		os.Exit(1)
	}
	fmt.Println("Key read.")
	// request the user's username
	fmt.Print("Username: ")
	var username string
	fmt.Scanf("%s", &username)
	// request the user's color
	fmt.Print("Color (leave blank for gray): ")
	var color string
	fmt.Scanf("%s", &color)
	switch strings.ToLower(color) {
	case "":
		color = peerutils.Gray
	case "white":
		color = peerutils.White
	case "red":
		color = peerutils.Red
	case "blue":
		color = peerutils.Blue
	case "green":
		color = peerutils.Yellow
	case "magenta":
		color = peerutils.Magenta
	case "cyan":
		color = peerutils.Cyan
	case "yellow":
		color = peerutils.Yellow
	default:
		fmt.Println("Invalid color")
		os.Exit(1)
	}
	// generate the user's ID
	id, err := peerutils.GenId(&prvKey)
	if err != nil {
		fmt.Printf("%vError:%v Failed to generate a user id. Exiting...\n", peerutils.Red, peerutils.ColorReset)
	}
	user := peerutils.User{Name: username, Color: color, Id: id}
	return prvKey, pubKey, user

}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("This program accepts exactly one argument.")
		return
	}
	command := os.Args[1]
	switch command {
	case "init":
		// request a path and generate and RSA key pair at that path
		fmt.Print("Key path: ")
		var keyPath string
		fmt.Scanf("%s", &keyPath)
		// create the keys
		fmt.Println("Generating RSA keys...")
		err := cryptoutils.GenerateRsaKeys(keyPath)
		if err != nil {
			fmt.Println("Failed to generate the keys. Does the key path exist?")
			return
		}
		fmt.Println("Keys generated succesfully")

	case "connect":
		prvKey, pubKey, user := login()
		var addr string
		fmt.Print("Peer address: ")
		fmt.Scanf("%s", &addr)
		fmt.Println("Connecting...")
		tunnel, err := peerutils.ConnectPeer(addr, pubKey, prvKey, user)
		if err != nil {
			fmt.Printf("Error: %v%v%v\n", peerutils.Red, err.Error(), peerutils.ColorReset)
			return
		}
		chatInterface := cliutils.NewChatInterface(tunnel)
		chatInterface.Run()

	case "await":
		prvKey, pubKey, user := login()
		fmt.Println("Listening...")
		tunnel, err := peerutils.AwaitPeer(pubKey, prvKey, user)
		if err != nil {
			fmt.Printf("Error: %v%v%v\n", peerutils.Red, err.Error(), peerutils.ColorReset)
			return
		}
		chatInterface := cliutils.NewChatInterface(tunnel)
		chatInterface.Run()

	default:
		fmt.Printf("Unrecognized command: %v\n", command)
	}
}
