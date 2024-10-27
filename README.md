# Courier
CourierCLI, an anonymous command-line messaging app. Messages are end-to-end encrypted using AES256 operating in GCM mode.
### Important note:
Because courier is peer-to-peer, and currently requires peers to know a user's IP address to connect, using a VPN or Proxy is **STRONGLY** reccomended when running Courier. 


# Roadmap/ToDo
- Add self-destructing messages
- Add group chats
- Add support for multiple chats at once
  
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
  - Run `courier init <keyPath>`
  - The keyPath is the relative path to the directory you'd like to store your RSA keys for courier. This will generate new directories as needed.
  - The directory will now contain two files: `pub.pem` and `prv.pem` which will be used as your public and private keys respectively.
  - As of version v0.1.2, you will be prompted to set a password to encrypt your private key with, it is recommended to set a password, as without one, your RSA private key will be stored in plaintext.

# Usage:
Once you've logged in, you'll be greeted with a prompt for a command. As of 10/16/2024, there are 3 available commands
- connect <address>
  - Takes a peer's IP address and attempts to connect to them
- await
  - Awaits incoming connections.
  - Note: In future versions, this command will be removed and will automatically run in the background
- clear:
  - Clears the screen
- read-archive <filepath>:
  - Prompts the user for the password for the archive file at `filepath` and displays the decrypted chat archive if the password is correct.
- exit:
  - Exits courier

## Logging in
When logging into courier you will recieve the following prompts
- Username:
  - This will be your display name. Can be anything you want as long as it is >= 64 characters in length.
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

### Current Commands
- clear:
  - Clears the chat history and clears all messages off screen
  - This only affects the client and has no effect on the peer's message history
- disconnect:
  - Terminates the connection with the peer
- peerid: 
    - Returns the peer's user ID. Because Courier is P2P with no centralized infrastructure, these IDs are the only way to verify a peer's identity, so it's important to ensure your peer has the ID you expect them to have.
- archive \<path> \[rounds]:
  - Creates a password-protected archive of the chat, and stores it to the specified directory (creating new directories as needed). The archive's file name is based on the current time, and it is name as `<HOUR>-<MINUTE>-<SECOND>.arc`
  - Specifiying the number of rounds is optional. Rounds determines the number of rounds of hashing your password will undergo to create an encryption key.
- delete \[id]
    - Deletes the message with the selected id
    - If no id is specified, it will default to the user's last-sent message
