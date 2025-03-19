package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type CommandHandler func(s *Server, conn net.Conn, payload []byte)

var commands = map[string]CommandHandler {
	"/quit":  		handleQuit,
	"/register": 	handleRegister,
	"/help":  		handleHelp,
	"/login": 		handleLogin,
	"/logout": 	 	handleLogout,
}

var commandDescriptions = [...]string {
	"- '/help': Lists all the available commands with a description.",
	"- '/quit': Signals the server to close the connection.",
	"- '/register <username> <password>': Sends username and locally hashed password to the server to set up a new user. If the given username is already in use an error will be returned.",
	"- '/login <username> <password>': Sends username and locally hashed password to the server to verify the combination of both. If valid you will be logged in. At 3 wrong login attempts this connection will be closed by the server.",
	"- '/logout': Logs you out of the user account you are currently logged in as.",
}

type Message struct {
	sender  net.Conn
	payload []byte
}

type Server struct {
	listenAddr 	   	string
	clientConns    	map[net.Conn]string
	clientConnsRev	map[string]net.Conn
	qtChs 		   	map[net.Conn]chan struct{}
	msgChannel	   	chan Message
	mu  		   	sync.Mutex
	usrPwdMap 	   	map[string]string
	muShadow 	   	sync.Mutex
}

func NewServer(listenAddr string) *Server {

	return &Server{
		listenAddr:   	listenAddr,
		clientConns:  	make(map[net.Conn]string),
		clientConnsRev:	make(map[string]net.Conn),
		qtChs:  	  	make(map[net.Conn]chan struct{}),
		msgChannel:   	make(chan Message),
		usrPwdMap: 	  	make(map[string]string),
	}

}

// Start starts up the server by spawning 2 goroutines in order
// to listen for incomming connections and process the input of the
// clients.
// It then blocks to receive a shutdown signal upon which
// the 2 goroutines will shut down and finally the programm will terminate.
func (s *Server) Start() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if !s.loadUserPasswordHashes() {
		fmt.Println("[Error] Loading passwords failed. Aborting...")
		return
	}

	wg.Add(2)
	go s.processMessageChannelInput(ctx, &wg)
	go s.acceptClientConnections(ctx, &wg)

	select {
	case <-sigCh:
		fmt.Println("[Log] Server Received shutdown signal. Initiating shutdown...")
		cancel()
	}

	wg.Wait()

	s.saveUserPasswordHashes()

	fmt.Println("[Log] Shutdown complete.")

}

// loadUserPasswordHashes reads from the shadow file which stores the
// username-hash pairs for login verification and loads those values
// into the servers usrPwdMap.
// If there is no shadow file yet, it will be created.
// Returns true on successfull loading and false otherwise
func (s *Server) loadUserPasswordHashes() bool {

	if !fileExists(shadowPath) {

		err := os.MkdirAll(serverDataDir, 0700)
		if err != nil {
			fmt.Println("[Error] Creating directory for server data:", err)
			return false
		}

		pwdFile, err := os.Create(shadowPath)
		if err != nil {
			fmt.Println("[Error] Creating shadow file:", err)
			return false
		}
		pwdFile.Close()
		fmt.Println("[Log] No shadow file yet. Created it.")
		return true
	}

	pwdFile, err := os.Open(shadowPath)
	if err != nil {
		fmt.Println("[Error] Opening shadow file:", err)
		return false
	}
	defer pwdFile.Close()

	pwdFileLen, err := getFileSize(shadowPath)
	if err != nil {
		fmt.Println("[Error] Gettings length of file:", err)
		return false
	}

	var buffer = make([]byte, pwdFileLen)
	_, err = pwdFile.Read(buffer)
	if err != nil {
		fmt.Println("[Error] Reading shadow file:" ,err)
		return false
	}

	shadowLines := strings.Split(string(buffer), "\n")
	for _, line := range shadowLines {

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry := strings.Split(line, ":")
		if len(entry) < 2 {
			fmt.Printf("[Warning] Invalid shadow file entry: %s\n", line)
			continue
		}
		s.usrPwdMap[entry[0]] = entry[1]

	}

	fmt.Println("[Log] Successfully loaded shadow file.")
	return true

}

// saveUserPasswordHashes writes all the entries from the servers
// usrPwdMap into a temporary shadow file. After that it overwrites
// the original shadow file.
// Returns true on success and false otherwise.
func (s *Server) saveUserPasswordHashes() bool {

	fmt.Println("[Log] Writing user-passwordHash pairs to temporary shadow file...")

	tempShadowFile, err := os.Create(tempShadowPath)
	if err != nil {
		fmt.Println("[Error] Creating tempShadowFile, aborting...:", err)
		return false
	}
	defer tempShadowFile.Close()

	// NOTE: Mutex muShadow not needed, because other threads have been canceled before.
	for user, pwdHsh := range s.usrPwdMap {

		line := user + ":" + pwdHsh + "\n"

		n, err := tempShadowFile.WriteString(line)
		if err != nil {
			fmt.Printf("[Error] Writing user-passwordHash pair of '%s' to temporary shadow file:\n%s\n", user, err)
			return false
		}
		if n != len(line) {
			fmt.Println("[Error] Couldn't write entire line into temporary shadow file.")
			return false
		}

	}

	fmt.Println("[Log] Moving temporary shadow file to permanent shadow file...")

	err = os.Rename(tempShadowPath, shadowPath)
	if err != nil {
		fmt.Println("[Error] Moving temporary shadow file to permanent shadow file.")
		return false
	}

	fmt.Println("[Log] Successfully saved user-passwordHash pairs.")

	return true

}

// acceptClientConnections starts a listener on the servers listenAddr.
// It then waits for new incomming connections and spawns a new goroutine
// per connection. If a cancellation signal is received via ctx, it waits 
// for all the summoned goroutines to terminate and then returns.
//
// Parameters:
// 	ctx - Context for cancellation of function
// 	mainWG - Waitgroup for syncing
func (s *Server) acceptClientConnections(ctx context.Context, mainWG *sync.WaitGroup) {

	defer mainWG.Done()

	var wg sync.WaitGroup

	// Start listener
	fmt.Println("[Log] Setting up listener...")
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		fmt.Println("[Error] Starting to listen for connections:", err)
		return
	}
	defer ln.Close()

	// Wait for incomming connections
	fmt.Println("[Log] Now listening for incomming client connections...")
	for {

		if err := ln.(*net.TCPListener).SetDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			fmt.Println("[Log] Error setting deadline for incomming client connections:", err)
			return
		}

		select {
		case <-ctx.Done():
			fmt.Println("[Log] Shutting down: Accepting incomming client connections...")
			fmt.Println("[Log] Waiting for client handlers to termiante...")
			wg.Wait()
			fmt.Println("[Log] All client handlers terminated.")
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Println("[Error] Accepting incomming connection:", err)
				continue
			}

			// Spawn client handler
			wg.Add(1)
			go s.handleClientConnection(ctx, &wg, conn)

		}
	}

}

// handleClientConnection sets up the client by adding the relevant
// information to the servers members. It then starts listeneing for
// any incomming data from the client. That data is then forwarded
// into the msgChannel to which the processMessageChannelInput
// function is listening.
//
// Parameters:
// 	ctx - Context for cancellation of function
// 	mainWG - Waitgroup for syncing
// 	conn - the connection to manage
func (s *Server) handleClientConnection(ctx context.Context, 
	 									wg *sync.WaitGroup, 
	 									conn net.Conn) {

	fmt.Println("[Log] Received new client. Setting up handler...")

	defer func() {
		s.mu.Lock()
		delete(s.clientConnsRev, s.clientConns[conn])
		delete(s.clientConns, conn)
		delete(s.qtChs, conn)
		s.mu.Unlock()
		conn.Close()
		wg.Done()
	}()

	s.mu.Lock()
	s.clientConns[conn] = "anonymous"
	s.qtChs[conn] 		= make(chan struct{})
	s.mu.Unlock()

	fmt.Println("[Log] New client is now set up.")

	b := make([]byte, 2048)
	for {

		err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if err != nil {
			fmt.Printf("[Error|%s] Setting up Read deadline:\n%s\n",conn.RemoteAddr(),  err)
			return
		}

		select {
		case <-ctx.Done():
			fmt.Println("[Log] Shutting down client handler...")
			return
		case <-s.qtChs[conn]:
			fmt.Println("[Log] Received '/quit' command. Shutting down connection.")
			return
		default:
			n, err := conn.Read(b)
			if err != nil {
				// checks if err implements the net.Error interface
				// if it does (ok = true), checks if a Timeout occured
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if err == io.EOF {
					fmt.Printf("[Log|%s] Client closed connection.\n", conn.RemoteAddr())
					return
				}
				fmt.Println("[Error] Reading message from client:", err)
				return
			}
			payload := b[:n]
			s.msgChannel <- Message{
				sender:  conn,
				payload: payload,
			}
		}

	}

}

// processMessageChannelInput listens to a channel into which every clients
// handler sends the received messages. It then prints out the content and calls
// the handler functions for the different commands.
//
// Parameters:
// 	ctx - Context for cancellation of function
// 	mainWG - Waitgroup for syncing
func (s *Server) processMessageChannelInput(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {
		case <-ctx.Done():
			fmt.Println("[Log] Shutting down: Processing of message channel...")
			return
		case msg :=<- s.msgChannel:
			fmt.Printf("[Log] Message received from %s:\n%s\n", msg.sender.RemoteAddr(), string(msg.payload))

			pld := string(msg.payload)
			if strings.HasPrefix(pld, "/") {
				fmt.Printf("[Log] Command from %s: '%s'.\n", msg.sender.RemoteAddr(), pld)
				command := strings.Fields(pld)[0]
				handler, ok := commands[command]
				if !ok {
					fmt.Printf("[Log] Command from %s was invalid: %s\n", msg.sender.RemoteAddr(), string(msg.payload))
					continue
				}

				handler(s, msg.sender, msg.payload)

			}
		}

	}

}

// sendMessageToClient sends the given message to the given client.
// It adds a prefix to the message to show the client as which user he is logged
// in as. In case of an error the given error message will be printed for context.
// To avoid race conditions the map of connections is blocked during the writing process.
//
// Parameters:
//	conn - the clients connection to send the message to
// 	msg - the message to send
//	errMsg - the error message to print for context
func (s *Server) sendMessageToClient(conn net.Conn, msg string, errMsg string) {

	username := "anonymous"

	s.mu.Lock()
	_, exists := s.clientConnsRev[s.clientConns[conn]]
	// fmt.Println("[Debugging] Current map:", s.clientConns)
	// fmt.Println("[Debugging] Current reverse map:", s.clientConnsRev)
	// fmt.Println("[Debugging] s.clientConns[conn] = ", s.clientConns[conn])
	// fmt.Println("[Debugging] s.clientConnsRev[s.clientConns[conn]] = ", s.clientConnsRev[s.clientConns[conn]])
	if exists {
		// fmt.Println("[Debugging] user is loggin in. Personalizing message...")
		username = s.clientConns[conn]
	}
	s.mu.Unlock()

	personalizedMsg := "(" + username + ") " + msg

	_, err := conn.Write([]byte(personalizedMsg))
	if err != nil {
		fmt.Println(errMsg + ":", err)
		return
	}
	return

}

// sendMessageToClientLocked sends the given message to the given client.
// It adds a prefix to the message to show the client as which user he is logged
// in as. In case of an error the given error message will be printed for context.
// The function assumes that the s.mu Mutex is locked. It can therefore be used
// in the context of a already locked Mutex.
//
// Parameters:
//	conn - the clients connection to send the message to
// 	msg - the message to send
//	errMsg - the error message to print for context
func (s *Server) sendMessageToClientLocked(conn net.Conn, msg string, errMsg string) {

	username := "anonymous"

	_, exists := s.clientConnsRev[s.clientConns[conn]]
	if exists {
		username = s.clientConns[conn]
	}

	personalizedMsg := "(" + username + ") " + msg

	_, err := conn.Write([]byte(personalizedMsg))
	if err != nil {
		fmt.Println(errMsg + ":", err)
		return
	}
	return

}

// -----------------------------
// ---------- Handler ----------
// -----------------------------

// handleQuit closes the channel of the given connection to terminate a connection.
//
// Parameters
// 	s - the server
// 	conn - the connection which will be closed
//	payload - the arguments of the command. Not used with this command.
func handleQuit(s *Server, conn net.Conn, payload []byte) {

	fmt.Printf("Handling '/quit' command from %s...\n", conn.RemoteAddr())

	close(s.qtChs[conn])

}

// handleRegister checks if a given username is already registered
// at the server. If it is not, a new user will be added to the server.
//
// Parameters:
// 	s - the server
// 	conn - the clients connection
// 	payload - the arguments of the command. In this case: <command> <username> <password-hash>
func handleRegister(s *Server, conn net.Conn, payload []byte) {

	fmt.Printf("Handling '/register' command from %s...\n", conn.RemoteAddr())

	slicedPld := strings.Fields(string(payload))
	username  := slicedPld[1]
	pwdHsh 	  := slicedPld[2]

	// fmt.Printf("[Debugging] Received username: %s, password hash: %s\n", username, pwdHsh)

	s.muShadow.Lock()
	_, exists := s.usrPwdMap[username]
	if exists || username == "anonymous" {
		fmt.Println("[Log] '/register' failed because of duplicate username.")
		s.sendMessageToClientLocked(conn, "[Error] Username already exists. Please retry with different username.", "")
		s.muShadow.Unlock()
		return
	}
	s.usrPwdMap[username] = pwdHsh
	s.muShadow.Unlock()

	fmt.Println("[Log] Successfully added new user to usrPwdMap.")

	msg    := "A new user has been added: " + username
	errMsg := "[Error|" + conn.RemoteAddr().String() + "] Writing 'new user added' message."
	s.sendMessageToClient(conn, msg, errMsg)

}

// handleHelp sends a list of all possible commands to the given user.
// 
// Parameters:
// 	s - the server
// 	conn - the clients connection
// 	payload - the arguments of the command. Not used with this command.
func handleHelp(s *Server, conn net.Conn, payload []byte) {

	fmt.Printf("Handling '/help' command from %s...\n", conn.RemoteAddr())

	var builder strings.Builder
	builder.Write([]byte("The following is a list of all available commands and their usecase:\n"))
	for _, line := range commandDescriptions {
		_, err := builder.Write([]byte(line + "\n"))
		if err != nil {
			fmt.Println("[Error] Concatenation of command description failed:", err)

			msg := "[Error] Something went wrong at the server. Please try again..."
			errMsg := "[Error] Writing 'failed concatenation' message to the client:"
			s.sendMessageToClient(conn, msg, errMsg)

			return
		}
	}

	msg    := builder.String()
	errMsg := "[Error] Writing list of command descriptions to " + conn.RemoteAddr().String()
	s.sendMessageToClient(conn, msg, errMsg)

}

// handleLogin maps a connection onto a user, thery logging him in.
// The function checks if the user with the given username is already logged
// in to avoid duplicate log ins. Invalid usernames and passwords are
// also checked. If valid credentials are provided and the user isn't already
// logged in, the servers maps are updated, to map the connection onto the user.
//
// Parameters:
// 	s - the server
// 	conn - the clients connection
// 	payload - the arguments of the command. In this case: <command> <username> <password-hash>
func handleLogin(s *Server, conn net.Conn, payload []byte) {

	fmt.Printf("Handling '/login' command from %s...\n", conn.RemoteAddr())

	slicedPld  	  := strings.Fields(string(payload))
	inputUsername := slicedPld[1]
	inputPwdHsh   := slicedPld[2]

	// fmt.Printf("[Debugging] Received username: %s, password hash: %s as a login combination.\n", inputUsername, inputPwdHsh)

	// Handle duplicate login
	s.mu.Lock()
	_, userLoggedIn := s.clientConnsRev[inputUsername]
	if userLoggedIn {
		fmt.Printf("[Log] '/login'failed because user '%s' was already logged in.\n", inputUsername)
		msg    := "[Error] Login failed because user is already logged in."
		errMsg := "[Error] Failed writing 'duplicate login' message to " + conn.LocalAddr().String()
		s.sendMessageToClientLocked(conn, msg, errMsg)

		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	s.muShadow.Lock()
	// Handle invalid username
	pwdHsh, userExists := s.usrPwdMap[inputUsername]
	if !userExists {
		fmt.Println("[Log] '/login' failed because invalid username was given.")
		msg    := "[Error] Invalid combiation of username and password given."
		errMsg := "[Error] Failed writin 'invalid username' message to " + conn.LocalAddr().String()
		s.sendMessageToClientLocked(conn, msg, errMsg)	

		s.muShadow.Unlock()
		return
	}

	// fmt.Printf("[Debugging] shadowfileHash: %s\ninputHash: %s\n", pwdHsh, inputPwdHsh)

	// Handle wrong password
	if pwdHsh != inputPwdHsh {
		fmt.Println("[Log] '/login' failed because invalid password hash was given.")
		msg    := "[Error] Invalid combiation of username and password given."
		errMsg := "[Error] Writin 'wrong password' message to " + conn.RemoteAddr().String()
		s.sendMessageToClient(conn, msg, errMsg)

		s.muShadow.Unlock()
		return
	}
	s.muShadow.Unlock()

	s.mu.Lock()
	s.clientConns[conn] = inputUsername
	s.clientConnsRev[inputUsername] = conn
	s.mu.Unlock()

	msg    := "Login successfull." 
	errMsg := "[Error] Failed writing 'successfull login' message to " + conn.RemoteAddr().String()
	s.sendMessageToClient(conn, msg, errMsg)

}

func handleLogout(s *Server, conn net.Conn, payload []byte) {

	// TODO: Send success-message to client
	fmt.Printf("Handling '/logout' command from %s...\n", conn.RemoteAddr())

	s.mu.Lock()
	delete(s.clientConnsRev, s.clientConns[conn])
	s.clientConns[conn] = "anonymous"
	s.mu.Unlock()

	msg := "Logout successfully."
	errMsg := "[Error] Failed writing 'logout successfull' message to " + conn.RemoteAddr().String()
	s.sendMessageToClient(conn, msg, errMsg)

	fmt.Println("[Log] Successfully logged out", conn.RemoteAddr())

}
