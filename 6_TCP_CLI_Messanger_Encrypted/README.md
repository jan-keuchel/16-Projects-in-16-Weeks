# TCP CLI Messanger Encrypted

## Goal
The goal of this project is to create a messanger application which is based in the terminal.
Using this application it is possible to connect to the server, register new users and log in as a user. From there on one can create new chats with other users aswell as group-chats. The goal is for the server not to know, what its clients are writing about. Meaning, every message will be encrypted at the clients machine, then sent to the server, which will forward the message to the respective client, where the message will be decrypted again.
At first the login will happen via username-password pairs. Later on that will maybe change to a public-private-key based authentification.
As an improvement on the last project, another goal is to close conections properly and not just shut them down. 

## Features
- [ ] Managing connections
    - [x] The server listens for new connections indefinetely
    - [ ] The connections can be closed by the client ('/quit') or by the server (3 wrong login attempts)
- [ ] Commands
    - [x] '/quit' - logs out the client and closes the connection
    - [x] '/login' - initiates the login process
        - [x] Query for username (locally)
        - [x] Query for passowrd (locally)
    - [x] '/logout' - logs out the user
    - [ ] '/register' - initiates the sign up process
        - [x] Query for username (locally) - No duplicate usernames
        - [x] query for passowrd (locally)
        - [ ] Generate public-private key-pair
    - [x] '/newChat \<username\>' - Sends a chat request to the user specified
        - [ ] The new chat will be assigned an ID
    - [ ] '/listChats' - lists the IDs and names of recipients of every chat
    - [ ] '/chat \<ID\>' - initiates switch to chat mode
        - [ ] retrieves the content of the chat, decrypts it and prints it to the screen
        - [ ] If the request is still pending, it is not possible to write messages
        - [ ] If accepted, the CLI now takes input as chat messages
        - [ ] It's possible to write messages to a client who is offline
    - [ ] '/exit' - exits chat mode and returns to overview
    - [ ] '/deleteChat \<id\>' - deletes a chat 
    - [x] '/help' - prints a list of commands along with their descriptions
- [ ] User Experience
    - [ ] Proper walk through of how to establish the connection

## Open questions
- [ ] How do I encrypt and decrypt the messages locally on the client?
    - Symmetric encryption with Diffie-Hellman key exchange
- [ ] Do I - locally -  store one key per client I want to write to?
    - Yes, one key per chat
- [ ] How do I distribute keys in a group chat?
    - Diffie-Hellman key exchange

## Reflection and improvements

