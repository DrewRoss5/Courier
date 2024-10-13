package peerutils

import (
	"errors"
	"fmt"
	"os/exec"
)

const MAX_MSG_COUNT = 50

type Chatroom struct {
	Tunnel   Tunnel
	Messages []Message
	Active   bool
}

// appends a new message to the chat history
func (c *Chatroom) pushMessage(msg *string, user *User) {
	newMessage := NewMessage(*msg, user)
	c.Messages = append(c.Messages, *newMessage)
	// delete old Messages if the archive is too large TODO: Make this an optional feature and not mandated
	if len(c.Messages) > MAX_MSG_COUNT {
		c.Messages = c.Messages[1:]
	}
}

// awaits an incoming message, and handles it according to its code
func (c *Chatroom) AwaitMessage() error {
	if !c.Active {
		return errors.New("chatroom: this chatroom is no longer Active")
	}
	msg, err := c.Tunnel.AwaitMessage()
	if err != nil {
		return err
	}
	// seperate the message from its code and handle it accordingly
	msgCode := msg[0]
	msg = msg[1:]
	switch msgCode {
	case MESSAGE_TXT:
		fmt.Println("Message recieved")
		messageStr := string(msg)
		c.pushMessage(&messageStr, &c.Tunnel.Peer)
	case MESSAGE_DISCONNECT:
		c.Active = false
	default:
		return errors.New("chatroom: invalid message recieved")
	}
	return nil
}

// sends a string message to the peer
func (c *Chatroom) SendMessage(msg *string) error {
	tmp := []byte(*msg)
	msgBytes := append([]byte{MESSAGE_TXT}, tmp...)
	err := c.Tunnel.SendMessage(msgBytes)
	if err != nil {
		return err
	}
	c.pushMessage(msg, &c.Tunnel.User)
	return nil

}

// displays all of the Messages currently in the archive
func (c Chatroom) DisplayMessages() {
	if len(c.Messages) == 0 {
		fmt.Printf("%vNo messages to display%v\n", Gray, ColorReset)
		return
	}
	for _, msg := range c.Messages {
		msg.Display()
	}
}

// handles a chat command, and returns the string to be output to the terminal after running
func (c *Chatroom) HandleCommand(command string, args []string) string {
	switch command {
	case ">clear":
		clear(c.Messages)
		exec.Command("clear")
		return fmt.Sprintf("%vMessages cleared%v\n", Gray, ColorReset)
	case ">disconnect":
		err := c.Tunnel.Shutdown()
		if err != nil {
			return fmt.Sprintf("%vError: failed to close the chatroom%v\n", Red, ColorReset)
		}
		return fmt.Sprintf("%vChat closed.%v\n", Gray, ColorReset)
	default:
		return fmt.Sprintf("%vError: unrecognized command%v\n", Red, ColorReset)
	}
}
