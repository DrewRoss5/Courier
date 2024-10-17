package main

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/DrewRoss5/courier/cliutils"
	"github.com/DrewRoss5/courier/cryptoutils"

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
	prvKey, pubKey, user := login()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%v%v%v > ", user.Color, user.Name, peerutils.ColorReset)
		input, _ := reader.ReadString('\n')
		tmp := strings.Split(input, " ")
		command := strings.Replace(tmp[0], "\n", "", 1)
		commandArgs := tmp[1:]
		switch command {
		case "await":
			// await an incoming connection and run the chatroom
			fmt.Println("Listening...")
			tunnel, err := peerutils.AwaitPeer(pubKey, prvKey, user)
			if err != nil {
				fmt.Printf("%verror:%v %v\n", peerutils.Red, peerutils.ColorReset, err.Error())
				continue
			}
			chat := cliutils.NewChatInterface(tunnel)
			chat.Run()
		case "connect":
			if len(commandArgs) != 1 {
				fmt.Printf("%verror:%v This command takes exactly one argument\n", peerutils.Red, peerutils.ColorReset)
				continue
			}
			fmt.Println("Connecting...")
			addr := strings.Replace(string(commandArgs[0]), "\n", "", 1)
			tunnel, err := peerutils.ConnectPeer(addr, pubKey, prvKey, user)
			if err != nil {
				fmt.Printf("%verror:%v %v\n", peerutils.Red, peerutils.ColorReset, err.Error())
				continue
			}
			chat := cliutils.NewChatInterface(tunnel)
			chat.Run()
		case "clear":
			// determine if we're running on windows, which uses a different clear command
			clearCommand := "clear"
			if runtime.GOOS == "windows" {
				clearCommand = "cls"
			}
			cmd := exec.Command(clearCommand)
			cmd.Stdout = os.Stdout
			cmd.Run()
		default:
			fmt.Printf("%verror:%v Unrecognized command\n", peerutils.Red, peerutils.ColorReset)
		}

	}
}
