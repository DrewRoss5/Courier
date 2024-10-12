package peerutils

import "errors"

const MAX_MSG_COUNT = 50

type Chatroom struct {
	tunnel   tunnel
	messages []*message
	active   bool
}

// appends a new message to the chat history
func (c *Chatroom) pushMessage(msg *string, user *User) {
	newMessage := NewMessage(*msg, user)
	c.messages = append(c.messages, newMessage)
	// delete old messages if the archive is too large TODO: Make this an optional feature and not mandated
	if len(c.messages) > MAX_MSG_COUNT {
		c.messages = c.messages[1:]
	}
}

// awaits an incoming message, and handles it according to its code
func (c *Chatroom) AwaitMessage() error {
	if !c.active {
		return errors.New("chatroom: this chatroom is no longer active")
	}
	msg, err := c.tunnel.AwaitMessage()
	if err != nil {
		return err
	}
	// seperate the message from its code and handle it accordingly
	msgCode := msg[0]
	msg = msg[1:]
	switch msgCode {
	case MESSAGE_TXT:
		messageStr := string(msg)
		c.pushMessage(&messageStr, &c.tunnel.Peer)
	case MESSAGE_DISCONNECT:
		c.active = false
	}
	return nil
}

// sends a string message to the peer
func (c *Chatroom) SendMessage(msg *string) error {
	tmp := []byte(*msg)
	msgBytes := append([]byte{MESSAGE_TXT}, tmp...)
	err := c.tunnel.SendMessage(msgBytes)
	if err != nil {
		return err
	}
	c.pushMessage(msg, &c.tunnel.User)
	return nil

}

// displays all of the messages currently in the archive
func (c Chatroom) DisplayMessages() {
	for _, msg := range c.messages {
		msg.Display()
	}
}
