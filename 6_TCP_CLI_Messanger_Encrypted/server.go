package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Message struct {
	sender  net.Conn
	payload []byte
}

type Server struct {
	listenAddr 	 string
	quitChannels map[net.Conn]<-chan bool
	clientConns  map[net.Conn] bool
	msgChannel	 chan Message
	mu  		 sync.Mutex
}

func NewServer(listenAddr string) *Server {

	return &Server{
		listenAddr:   listenAddr,
		quitChannels: make(map[net.Conn]<-chan bool),
		clientConns:  make(map[net.Conn]bool),
		msgChannel:   make(chan Message),
	}

}

func (s *Server) Start() {

	go s.processMessageChannelInput()
	s.acceptClientConnections()

}

// acceptClientConnections starts a listener on the servers specified address and
// accepts incomming client connections. The accepted client connection is then
// handled by handleClientConnection in a separate goroutine.
//
// The following operations are performed:
// 	 - Starting up the listener
// 	 - Waiting for and accepting incomming connections
// 	 - Starting a new goroutine per incomming connection
func (s *Server) acceptClientConnections() {

	fmt.Println("[Log] Setting up listener...")
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		fmt.Println("[Error] Starting to listen for connections:", err)
		return
	}
	defer ln.Close()

	fmt.Println("[Log] Now listening for incomming client connections...")
	for {

		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[Error] Accepting incomming connection:", err)
			continue
		}
		go s.handleClientConnection(conn)

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
func (s *Server) handleClientConnection(conn net.Conn) {

	quit := make(<-chan bool)

	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	s.mu.Lock()
	s.quitChannels[conn] = quit
	s.clientConns[conn] = true
	s.mu.Unlock()

	b := make([]byte, 2048)
	for {

		select {
		case <-quit:
			fmt.Printf("[Log|%s] Shutting down client handler...", conn.RemoteAddr())
			return
		default:
			n, err := conn.Read(b)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("[Log|%s] Client closed connection.\n", conn.RemoteAddr())
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

func (s *Server) processMessageChannelInput() {

	for msg := range s.msgChannel {

		fmt.Printf("[Log] Message received from %s:\n%s\n", msg.sender.RemoteAddr(), string(msg.payload))

	}

}
