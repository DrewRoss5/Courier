package peerutils

import (
	"fmt"
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

type message struct {
	content  string
	timeSent string
	sender   *User
}

func (m message) Display() {
	fmt.Printf("%v%v\033[0m@%v: %v\n", m.sender.Color, m.sender.Name, m.timeSent, m.content)
}

// constructs a new message, making a note of the current timestamp
func NewMessage(content string, usr *User) *message {
	// get the current time in minutes and seconds
	now := time.Now()
	timeStr := fmt.Sprintf("%v:%v", now.Hour(), now.Minute())
	return &message{content: content, sender: usr, timeSent: timeStr}
}
