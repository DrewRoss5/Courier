package peerutils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/DrewRoss5/courier/cryptoutils"
	"golang.org/x/term"
)

const (
	DEFAULT_ROUNDS = 256
	SALT_SIZE      = 16
	MAX_MSG_COUNT  = 50
)

type Chatroom struct {
	Tunnel   Tunnel
	Messages map[uint32]Message
	MaxId    uint32
	Active   bool
	Mut      sync.Mutex
}

// appends a new message to the chat history
func (c *Chatroom) pushMessage(msg *string, user *User) {
	newMessage := NewMessage(*msg, user)
	c.Mut.Lock()
	c.Messages[c.MaxId] = *newMessage
	c.MaxId++
	if c.MaxId == 0xffffffff {
		c.Messages = make(map[uint32]Message)
		c.serverMessage("Message limit reached. Chat history cleared")
	}
	c.Mut.Unlock()
}

// appends a new message sent from the chatroom itself, used for alerts, and the like
func (c *Chatroom) serverMessage(msg string) {
	colored := fmt.Sprintf("%v%v%v\n", Green, msg, ColorReset)
	c.pushMessage(&colored, &User{Name: "Chatroom", Id: "", Color: Green})
}

// pushes an error message to chatroom
func (c *Chatroom) errorMessage(msg string) {
	colored := fmt.Sprintf("%v%vError:%v %v %v\n", Bold, Red, Yellow, msg, ColorReset)
	c.pushMessage(&colored, &User{Name: "Chatroom", Id: "", Color: Green})
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
		messageStr := string(msg)
		c.pushMessage(&messageStr, &c.Tunnel.Peer)
	case MESSAGE_DISCONNECT:
		c.Active = false
	case MESSAGE_DELETE:
		id := binary.LittleEndian.Uint32(msg)
		message, ok := c.Messages[id]
		if ok && message.sender.Id == c.Tunnel.Peer.Id {
			c.Mut.Lock()
			delete(c.Messages, id)
			c.Mut.Unlock()
		}
	case CHAT_ARCHIVE:
		c.serverMessage(fmt.Sprintf("%v%v%v archived this chat.", c.Tunnel.Peer.Color, c.Tunnel.Peer.Name, Green))
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

// sends a message and waits for a given number of seconds before deleting it
func (c *Chatroom) TimedMessage(msg *string, delay int) {
	id := c.MaxId
	err := c.SendMessage(msg)
	if err != nil {
		c.Active = false
		return
	}
	time.Sleep(time.Second * time.Duration(delay))
	c.DeleteMessage(id)
}

// deletes the message with a specified ID from the chat
func (c *Chatroom) DeleteMessage(id uint32) {
	c.Mut.Lock()
	delete(c.Messages, id)
	c.Mut.Unlock()
	buf := make([]byte, 4)
	binary.LittleEndian.AppendUint32(buf, id)
	err := c.Tunnel.SendMessage(append([]byte{MESSAGE_DELETE}, buf...))
	if err != nil {
		c.Active = false
	}
}

// displays all of the Messages currently in the archive
func (c *Chatroom) DisplayMessages(file io.Writer) {
	if len(c.Messages) == 0 {
		fmt.Printf("%v%vNo messages to display%v\n", Italic, Gray, ColorReset)
		return
	}
	var i uint32
	for i = 0; i < c.MaxId; i++ {
		// under normal circumstances, all id's should be sequential, however, this is to account for messages potentially being deleted
		message, ok := c.Messages[i]
		if ok {
			message.Display(file)
		}
	}
}

// handles a chat command, and returns the string to be output to the terminal after running
func (c *Chatroom) HandleCommand(command string, args []string) {
	switch command {
	// clears the message history
	case ">clear":
		c.Messages = make(map[uint32]Message)
		c.serverMessage("Messages cleared")
	// terminates the connection
	case ">disconnect":
		err := c.Tunnel.Shutdown()
		if err != nil {
			c.errorMessage("failed to close the chatroom")
			return
		}
		c.serverMessage("Chat closed.")
	// returns the peer's ID
	case ">peerid":
		c.serverMessage(fmt.Sprintf("%v%v%v Has the ID\n%v%v%v", c.Tunnel.Peer.Color, c.Tunnel.Peer.Name, ColorReset, Green, c.Tunnel.Peer.Id, ColorReset))
	// deletes a message from the chat history on both user's ends
	case ">delete":
		var id uint32
		if len(args) == 0 {
			id = c.MaxId - 1
			// find the last message the user sent
			for id > 0 {
				message, ok := c.Messages[id]
				if ok && message.sender.Id == c.Tunnel.User.Id {
					break
				}
				id--
			}
		} else {
			tmp, err := strconv.Atoi(args[0])
			if err != nil {
				c.errorMessage("Invalid message id")
				return
			}
			id = uint32(tmp)
		}
		message, ok := c.Messages[id]
		if !ok || message.sender.Id != c.Tunnel.User.Id {
			c.errorMessage("Invalid message id")
			return
		}
		c.DeleteMessage(id)
	// sends a timed message that will automatically delete
	case ">timed":
		if len(args) < 2 {
			c.errorMessage("This command takes at least two arguments")
			return
		}
		delay, err := strconv.Atoi(args[0])
		if err != nil {
			c.errorMessage("Invalid delay time")
			return
		}
		msg := strings.Join(args[1:], " ")
		msg += "\n"
		go c.TimedMessage(&msg, delay)
	case (">color"):
		if len(args) < 2 {
			c.errorMessage("This command takes at least two arguments")
		}
		color, err := ParseColor(args[0])
		if err != nil {
			c.errorMessage("Invalid color")
			return
		}
		// send the colored message
		msg := color
		msg += strings.Join(args[1:], " ")
		msg += ColorReset + "\n"
		c.SendMessage(&msg)

	// archive the chat
	case ">archive":
		if len(args) < 1 || len(args) > 2 {
			c.errorMessage("That command takes between one and two arguments")
			return
		}
		// determine the rounds of hashing the user would like to use, DEFAULT=256
		var rounds int
		var err error = nil
		if len(args) == 1 {
			rounds = DEFAULT_ROUNDS
		} else {
			rounds, err = strconv.Atoi(args[1])
			if err != nil {
				c.errorMessage("Rounds must be a positive integer")
				return
			}
			if rounds < 0 {
				c.errorMessage("Rounds must be a positive integer")
				return
			}
		}
		fmt.Print("Password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			c.errorMessage("Failed to read the message")
			return
		}
		fmt.Print("\nConfirm Password: ")
		confirm, _ := term.ReadPassword(int(syscall.Stdin))
		if slices.Compare(confirm, password) != 0 {
			c.errorMessage("Password does not match confirmation. Archive not saved")
			return
		}
		err = c.ArchiveChat(password, args[0], rounds)
		if err != nil {
			c.errorMessage(err.Error())
			return
		}
		// inform the other user that the chat has been archived
		err = c.Tunnel.SendMessage([]byte{CHAT_ARCHIVE})
		if err != nil {
			c.Active = false
			c.errorMessage("connection severed")
			return
		}
		c.serverMessage("Chat archived")
	// severs the connection and exits the application
	case ">exit":
		err := c.Tunnel.Shutdown()
		if err != nil {
			c.errorMessage("failed to close the chatroom")
			return
		}
		os.Exit(0)
	default:
		c.errorMessage("unrecognized message")
		return
	}
}

// archives a chat and saves it to a file in a given path
func (c *Chatroom) ArchiveChat(password []byte, path string, rounds int) error {
	// create the path if needed
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}
	// generate a key
	salt := cryptoutils.GenNonce()
	key := cryptoutils.HashKey(password, salt, rounds)
	// write the chat to a buffer, and encrypt it
	var buf bytes.Buffer
	c.DisplayMessages(&buf)
	ciphertext, err := cryptoutils.AesEncrypt(buf.Bytes(), key)
	if err != nil {
		return err
	}
	// parse the rounds into a buffer
	roundBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(roundBuf, uint32(rounds))
	// save the salt and ciphertext to the file
	contents := append(roundBuf, salt...)
	contents = append(contents, ciphertext...)
	fileName := time.Now().Format("/15-04-05.arc")
	file, err := os.Create(path + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(contents)
	return nil
}

// returns a string of a decrypted chat archive
func DecryptArchive(fileName string, password []byte) (string, error) {
	// read and parse the file
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	if len(fileContents) < cryptoutils.AES_MIN_CIPHERTEXT_SIZE+SALT_SIZE {
		return "", errors.New("archive: invalid archive file")
	}
	roundsBuf := fileContents[:4]
	salt := fileContents[4 : SALT_SIZE+4]
	ciphertext := fileContents[SALT_SIZE+4:]
	// read the rounds
	rounds := binary.LittleEndian.Uint32(roundsBuf)
	// generate the key and attempt to decrypt the archive
	key := cryptoutils.HashKey(password, salt, int(rounds))
	archive, err := cryptoutils.AesDecrypt(ciphertext, key)
	if err != nil {
		return "", err
	}
	return string(archive), nil
}
