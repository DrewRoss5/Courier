# Courier
CourierCLI, an anonymous command-line messaging app. Messages are end-to-end encrypted using AES256 operating in GCM mode.

# Roadmap/ToDo
- Add an in-app command pallete for recieving/creating connections
- Add more robust error handling
- Implement support for password protecting private keys
- Allow for archiving messages
- Add additional chat commands
- Allow for multiple chats simulataneously (each chat having its own "tab")
- Add support for Windows

# Installation and Setup
## Linux Installation
- At the moment, there is no offical install script, so the app must be compiled manually
- Ensure you have [golang](https://go.dev/doc/install) installed
- Clone this repo
- Run the following command in this repo's directory:
  - `go build .`
- The directory will now contain a binary named `courier`. This is the executable for the Courier application. Run it from the command line to use the app.
- Optional: Run the following command to allow for calling courier anywhere
  - `sudo mv courier /usr/bin/courier`

## Setup: 
- Courier uses RSA encryption and signing for user veirification and key exchanges.
- To create a new RSA key pair:
  - Create a new empty directory to store the keys in
  - Run `courier init`
  - When prompted for a "keyPath" enter the relative path to the directory you just created
  - The directory will now contain two files: `pub.pem` and `prv.pem` which will be used as your public and private keys respectively.

# Usage:
At the moment, courier is launched by using one of two commands (this will change in future versions)
`courier await`: Awaits incoming messaging requests
`courier connect`: Connects to a peer, given their address

## Logging in
When logging into courier you will recieve the following prompts
- Username:
  - This will be your display name. Can be anything you want
- Color: 
  - This will be the color your name is displayed in 
  - Valid options:
    - Red
    - Green
    - Blue
    - Yellow
    - Cyan
    - Magenta 
    - White
- Key path:
  - The path to the RSA key pair you'll be using for key transmission/message signing

## Commands
Because courier is a CLI application, interaction other than sending messages is done via commands.
To use a command, type ">" in the message entry immediately followed by the command (no space).
More commands will be added in future builds

### Current Commands
- Clear:
  - Clears the chat history and clears all messages off screen
  - This only affects the client and has no effect on the peer's message history
- Disconnect:
  - Terminates the connection with the peer
- Peerid: 
    - Returns the peer's user ID. Because Courier is P2P with no centralized infrastructure, these IDs are the only way to verify a peer's identity, so it's important to ensure your peer has the ID you expect them to have. 
