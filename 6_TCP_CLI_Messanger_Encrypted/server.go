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
}

type Message struct {
	sender  net.Conn
	payload []byte
}

type Server struct {
	listenAddr 	 string
	clientConns  map[net.Conn]bool
	qtChs 		 map[net.Conn]chan struct{}
	msgChannel	 chan Message
	mu  		 sync.Mutex
	usrPwdMap 	 map[string]string
	shadowNu 	 sync.Mutex
}

func NewServer(listenAddr string) *Server {

	return &Server{
		listenAddr:   listenAddr,
		clientConns:  make(map[net.Conn]bool),
		qtChs:  	  make(map[net.Conn]chan struct{}),
		msgChannel:   make(chan Message),
		usrPwdMap: 	  make(map[string]string),
	}

}

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

	fmt.Println("[Log] Shutdown complete.")

}

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

// acceptClientConnections starts a listener on the servers specified address and
// accepts incomming client connections. The accepted client connection is then
// handled by handleClientConnection in a separate goroutine.
//
// The following operations are performed:
// 	 - Starting up the listener
// 	 - Waiting for and accepting incomming connections
// 	 - Starting a new goroutine per incomming connection
func (s *Server) acceptClientConnections(ctx context.Context, mainWG *sync.WaitGroup) {

	defer mainWG.Done()

	var wg sync.WaitGroup

	fmt.Println("[Log] Setting up listener...")
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		fmt.Println("[Error] Starting to listen for connections:", err)
		return
	}
	defer ln.Close()

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

			wg.Add(1)
			go s.handleClientConnection(ctx, &wg, conn)

		}
	}

}

// handleClientConnection manages a single client connection for the server.
// It handles reading messages from the client, forwarding them to the server's
// message channel, and properly cleaning up resources when the connection is closed.
//
// The function runs in its own goroutine for each client connection and performs
// the following key operations:
//   - Registers the client connection with the server
//   - Creates a quit channel for proper shutdown
//   - Reads incoming messages into a 2048-byte buffer
//   - Forwards messages to the server's message channel
//   - Handles connection errors and EOF conditions
//   - Ensures proper cleanup through defer statements
//
// Parameters:
//   conn - The net.Conn representing the client connection
//
// All connection state changes are protected by the server's mutex to ensure
// thread safety. Connection closure is logged with the client's remote address.
func (s *Server) handleClientConnection(ctx context.Context, 
	 									wg *sync.WaitGroup, 
	 									conn net.Conn) {

	fmt.Println("[Log] Received new client. Setting up handler...")

	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		delete(s.qtChs, conn)
		s.mu.Unlock()
		conn.Close()
		wg.Done()
	}()

	s.mu.Lock()
	s.clientConns[conn] = true
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

// -----------------------------
// ---------- Handler ----------
// -----------------------------

func handleQuit(s *Server, conn net.Conn, payload []byte) {

	fmt.Printf("Handling '/quit' command from %s...\n", conn.RemoteAddr())

	close(s.qtChs[conn])

}

func handleRegister(s *Server, conn net.Conn, payload []byte) {

	fmt.Printf("Handling '/register' command from %s...\n", conn.RemoteAddr())

	slicedPld := strings.Fields(string(payload))
	username  := slicedPld[1]
	pwdHsh 	  := slicedPld[2]

	fmt.Printf("[Debugging] Received username: %s, password hash: %s\n", username, pwdHsh)

	// TODO: user already exists?
	// TODO: Write new user to file

}
