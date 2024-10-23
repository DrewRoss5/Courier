package main

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/DrewRoss5/courier/cliutils"
	"github.com/DrewRoss5/courier/cryptoutils"
	"golang.org/x/term"

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
		color = peerutils.Green
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
	// generate the key file if requested
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if len(os.Args) != 3 {
			fmt.Printf("%verror:%v This command takes exactly one argument\n", peerutils.Red, peerutils.ColorReset)
			return
		}
		path := os.Args[2]
		err := os.MkdirAll(path, 0777)
		if err != nil {
			fmt.Printf("%verror:%v invalid path", peerutils.Red, peerutils.ColorReset)
			return
		}
		fmt.Println("Generating keys...")
		err = cryptoutils.GenerateRsaKeys(path)
		if err != nil {
			fmt.Printf("%verror:%v failed to generate the RSA keys", peerutils.Red, peerutils.ColorReset)
			return
		}
		fmt.Println("Key pair generated")
		return
	}
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
