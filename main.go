package main

import (
	"fmt"
	"os"
	"slices"
	"syscall"

	"github.com/DrewRoss5/courier/cliutils"
	"github.com/DrewRoss5/courier/cryptoutils"
	"github.com/DrewRoss5/courier/peerutils"
	"golang.org/x/term"
)

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
		fmt.Print("Private key password (leave blank to disable encryption):")
		password, _ := term.ReadPassword(int(syscall.Stdin))
		if len(password) == 0 {
			fmt.Printf("\n%vWarning:%v your private key will be stored in plaintext.\n", peerutils.Yellow, peerutils.ColorReset)
			password = nil
		} else {
			fmt.Print("Confirm: ")
			confirm, _ := term.ReadPassword(int(syscall.Stdin))
			if slices.Compare(password, confirm) != 0 {
				fmt.Printf("\n%verror:%v password does not match confirmation. Exiting...\n", peerutils.Red, peerutils.ColorReset)
				return
			}
		}
		fmt.Println("\nGenerating keys...")
		err = cryptoutils.GenerateRsaKeys(path, password)
		if err != nil {
			fmt.Printf("%verror:%v failed to generate the RSA keys", peerutils.Red, peerutils.ColorReset)
			return
		}
		fmt.Println("Key pair generated")
		return
	} else {
		cliutils.MainLoop()
	}
}
