package main

import (
	"fmt"
	"os"

	"github.com/DrewRoss5/courier/cliutils"
	"github.com/DrewRoss5/courier/cryptoutils"
	"github.com/DrewRoss5/courier/peerutils"
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
		fmt.Println("Generating keys...")
		err = cryptoutils.GenerateRsaKeys(path)
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
