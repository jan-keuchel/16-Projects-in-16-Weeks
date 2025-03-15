# TCP CLI Messanger Encrypted

## Goal
The goal is to have a CLI application in which someone can log into his account using username and password. Then he can retrieve chats with differend clients and write to those clients. It is also possible to create group chats.
The important part is that the server which stores the data of the chats should never see the actual messages. Meaning, all the data being sent should be encrypted locally at the client, then sent to the server and from there sent to the recipients who should then decrypt the message again.
The login an passwords will be stored similar to the '/etc/shadow' file in linux. 
It is also important to have properly managed client connections and close them properly (improve on mistakes from last project.)

## Features
- [ ] Managing connections
    - [ ] The server listens for new connections indefinetely
    - [ ] The connections can be closed by the client ('/quit') or by the server (3 wrong login attempts)
- [ ] Commands
    - [ ] '/quit' - logs out the client and closes the connection
    - [ ] '/login' - initiates the login process
        - [ ] Query for username (locally)
        - [ ] Query for passowrd (locally)
    - [ ] '/register' - initiates the sign up process
        - [ ] Query for username (locally) - No duplicate usernames
        - [ ] query for passowrd (locally)
    - [ ] '/newChat &ltusername&gt' - Sends a chat request to the user specified
        - [ ] The new chat will be assigned an ID
    - [ ] '/listChats' - lists the IDs and names of recipients of every chat
    - [ ] '/chat &ltID&gt' - initiates switch to chat mode
        - [ ] retrieves the content of the chat, decrypts it and prints it to the screen
        - [ ] If the request is still pending, it is not possible to write messages
        - [ ] If accepted, the CLI now takes input as chat messages
        - [ ] It's possible to write messages to a client who is offline
    - [ ] '/exit' - exits chat mode and returns to overview
    - [ ] '/deleteChat &ltid&gt' - deletes a chat 
    - [ ] '/help' - prints a list of commands along with their descriptions
- [ ] User Experience
    - [ ] Proper walk through of how to establish the connection

## Reflection and improvements

