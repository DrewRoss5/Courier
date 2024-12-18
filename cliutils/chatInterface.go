package cliutils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/DrewRoss5/courier/peerutils"
)

// a struct that provides an interface to chatrooms. In future versions, this will be very useful for managing multiple chats
type ChatInterface struct {
	room *peerutils.Chatroom
}

// clears the terminal and displays all messages
func (ci ChatInterface) Display() {
	// determine if we're running on windows, which uses a different clear command
	clearCommand := "clear"
	if runtime.GOOS == "windows" {
		clearCommand = "cls"
	}
	cmd := exec.Command(clearCommand)
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Printf("Chat with %v%v%v%v:\n", peerutils.Bold, ci.room.Tunnel.Peer.Color, ci.room.Tunnel.Peer.Name, peerutils.ColorReset)
	ci.room.DisplayMessages(os.Stdout)
}

// awaits a message and displays it (along with all other messages) once recieved
func (ci *ChatInterface) AwaitMessage() {
	for ci.room.Active {
		err := ci.room.AwaitMessage()
		ci.Display()
		fmt.Printf("%v%v%v%v: ", ci.room.Tunnel.User.Color, peerutils.Bold, ci.room.Tunnel.User.Name, peerutils.ColorReset)
		if err != nil {
			fmt.Printf("%vChat closed.\n%v", peerutils.Gray, peerutils.ColorReset)
			ci.room.Active = false
		}
	}
}

// awaits user input, and handles it if it's a command, or sends it if it is a message
func (ci *ChatInterface) AwaitInput() {
	fmt.Printf("%v%v%v%v: ", ci.room.Tunnel.User.Color, peerutils.Bold, ci.room.Tunnel.User.Name, peerutils.ColorReset)
	in := bufio.NewReader(os.Stdin)
	input, err := in.ReadString('\n')
	if err != nil {
		fmt.Printf("%vError: %v%v\n", peerutils.Red, err.Error(), peerutils.ColorReset)
		return
	}
	if ci.room.Active {
		// determine if the input is a command or a message, and handle it appropriately
		if input[0] == '>' {
			// run the input as a command
			input = strings.Replace(input, "\n", "", -1)
			tmp := strings.Split(input, " ")
			command := tmp[0]
			var args []string = nil
			if len(tmp) > 1 {
				args = tmp[1:]
			}
			ci.room.HandleCommand(command, args)
		} else {
			ci.room.SendMessage(&input)
		}
		ci.Display()
	}
}

// begins a chat session
func (ci *ChatInterface) Run() {
	ci.Display()
	go ci.AwaitMessage()
	for ci.room.Active {
		ci.AwaitInput()
	}
	fmt.Printf("%vConnection terminated%v\n", peerutils.Red, peerutils.ColorReset)
}

// initializes a ChatInterface, given the tunnel
func NewChatInterface(tunnel *peerutils.Tunnel) *ChatInterface {
	chatroom := peerutils.Chatroom{Tunnel: *tunnel, Active: true, MaxId: 0, Messages: make(map[uint32]peerutils.Message)}
	ci := ChatInterface{&chatroom}
	return &ci
}
