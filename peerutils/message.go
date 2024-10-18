package peerutils

import (
	"fmt"
	"io"
	"time"
)

const Red = "\033[31m"
const Green = "\033[32m"
const Yellow = "\033[33m"
const Blue = "\033[34m"
const Magenta = "\033[35m"
const Cyan = "\033[36m"
const Gray = "\033[37m"
const White = "\033[97m"
const ColorReset = "\033[0m"

type Message struct {
	content  string
	timeSent string
	sender   *User
}

// displays a message to an output stream, usually  this will be stdout, but this needs to be adjustable for archival purposes
func (m Message) Display(stream io.Writer) {
	fmt.Fprintf(stream, "%v%v%v @ %v%v%v: %v", m.sender.Color, m.sender.Name, ColorReset, Yellow, m.timeSent, ColorReset, m.content)
}

// constructs a new message, making a note of the current timestamp
func NewMessage(content string, usr *User) *Message {
	timeStr := time.Now().Format("15:04:05")
	return &Message{content: content, sender: usr, timeSent: timeStr}
}
