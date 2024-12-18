package cliutils

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/DrewRoss5/courier/cryptoutils"
	"github.com/DrewRoss5/courier/peerutils"
	"golang.org/x/term"
)

// there aren't real accounts, but this creates a user for "login"
func login() (rsa.PrivateKey, rsa.PublicKey, peerutils.User) {

	// read the user's key
	var keyPath string
	fmt.Print("Key path: ")
	fmt.Scanf("%s", &keyPath)
	fmt.Print("Private key password: ")
	keyPassword, _ := term.ReadPassword(syscall.Stdin)
	fmt.Println("\nImporting RSA keys...")
	prvKey, pubKey, err := cryptoutils.ImportRsa(keyPath, keyPassword)
	if err != nil {
		fmt.Println("Failed to import RSA keys. Exiting...")
		os.Exit(1)
	}
	fmt.Println("Keys read.")
	// request the user's username
	fmt.Print("Username: ")
	var username string
	fmt.Scanf("%s", &username)
	if len(username) > 64 {
		fmt.Printf("%vError:%v Invalid username\n", peerutils.Red, peerutils.ColorReset)
		os.Exit(1)
	}
	// request the user's color
	fmt.Print("Color (leave blank for gray): ")
	var color string
	fmt.Scanf("%s", &color)
	color, err = peerutils.ParseColor(color)
	if err != nil {
		fmt.Printf("%vError:%v Invalid color. Exiting...\n", peerutils.Red, peerutils.ColorReset)
		os.Exit(1)
	}
	// generate the user's ID
	id, err := peerutils.GenId(&prvKey)
	if err != nil {
		fmt.Printf("%vError:%v Failed to generate a user id. Exiting...\n", peerutils.Red, peerutils.ColorReset)
		os.Exit(1)
	}
	user := peerutils.User{Name: username, Color: color, Id: id}
	return prvKey, pubKey, user
}

func MainLoop() {
	// log the user in and begin the program loop
	prvKey, pubKey, user := login()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%v%v%v%v > ", peerutils.Bold, user.Color, user.Name, peerutils.ColorReset)
		input, _ := reader.ReadString('\n')
		input = strings.Replace(input, "\n", "", -1)
		tmp := strings.Split(input, " ")
		command := tmp[0]
		commandArgs := tmp[1:]
		switch command {
		case "await":
			// await an incoming connection and run the chatroomd
			fmt.Println("Listening...")
			tunnel, err := peerutils.AwaitPeer(pubKey, prvKey, user)
			if err != nil {
				fmt.Printf("%verror:%v %v\n", peerutils.Red, peerutils.ColorReset, err.Error())
				continue
			}
			chat := NewChatInterface(tunnel)
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
			chat := NewChatInterface(tunnel)
			chat.Run()
		case "read-archive":
			if len(commandArgs) != 1 {
				fmt.Printf("%verror:%v This command takes exactly one argument\n", peerutils.Red, peerutils.ColorReset)
				continue
			}
			fmt.Print("Password: ")
			password, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Printf("%vError:%v failed to read password\n", peerutils.Red, peerutils.ColorReset)
				continue
			}
			archive, err := peerutils.DecryptArchive(commandArgs[0], password)
			if err != nil {
				fmt.Printf("%vError:%v %v", peerutils.Red, peerutils.ColorReset, err.Error())
				continue
			}
			fmt.Printf("\n%v\n", archive)
		case "clear":
			// determine if we're running on windows, which uses a different clear command
			clearCommand := "clear"
			if runtime.GOOS == "windows" {
				clearCommand = "cls"
			}
			cmd := exec.Command(clearCommand)
			cmd.Stdout = os.Stdout
			cmd.Run()
		case "exit":
			return
		default:
			fmt.Printf("%verror:%v Unrecognized command\n", peerutils.Red, peerutils.ColorReset)
		}
	}
}
