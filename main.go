package main

import (
	"fmt"
	"os"

	"github.com/DrewRoss5/courier/cryptoutils"
)

func main() {
	// ensure the correct number of arguments
	if len(os.Args) != 3 {
		fmt.Println("This program accepts exactly two arguments ")
		return
	}
	//username := os.Args[1]
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

	case "initate":

	default:
		fmt.Println("Unrecognized command")
	}

}
