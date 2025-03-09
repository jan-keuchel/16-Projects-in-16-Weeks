package main

import (
	"fmt"
	"net"
	"sync"
)

type Message struct {
	sender  net.Conn
	payload []byte
}

type Server struct {
	listenAddr 	string
	clientConns map[net.Conn]bool
	ch  		chan Message
	mu 			sync.Mutex

}

// Returns a pointer to an initialized Server
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:  listenAddr,
		ch:  		 make(chan Message),
		clientConns: make(map[net.Conn]bool),
	}
}

// Receives data sent to the server via channel, calls function to process data and
// sends back the response for the clients
func (s *Server) broadcast() {

	for input := range s.ch {

		// TODO: handle client input
		fmt.Printf("[%s]: %s\n", input.sender.RemoteAddr().String(), string(input.payload))

	}

}

// Sets up a listener on specified port and waits for clients to connect. 
// Only accepts incomming connection if there are less than 2 clients connected.
func (s *Server) acceptClients() {

	fmt.Println("[Server] Setting up listener...")
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		fmt.Println("[Server] Error listening for client connections:", err)
		return
	}
	defer ln.Close()
	fmt.Println("[Server] Waiting for clients to connect...")

	for {

		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[Server] Error accepting incomming client connection:", err)
			continue
		}

		fmt.Println("[Server] New Connection request from client.")

		if len(s.clientConns) < 2 {
			go s.listenToClientConnection(conn)
		} else {
			fmt.Println("[Server] Connection request discarded: Already 2 clients.")
			_, err := conn.Write([]byte("Server is already full. Closing connection..."))
			if err != nil {
				fmt.Println("[Server] Error writing rejection due to full server:", err)
				continue
			}
			conn.Close()
		}

	}

}

// Adds connection (client) to the map of clients of the server, indefinitely
// reads from TCP connection and sends the read data into the servers channel.
func (s *Server) listenToClientConnection(conn net.Conn) {

	fmt.Println("[Server] Adding client to server...")

	defer conn.Close()
	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		s.mu.Unlock()
	}()

	s.mu.Lock()
	s.clientConns[conn] = true
	s.mu.Unlock()

	fmt.Println("[Server] New Client:", conn.RemoteAddr().String())

	greeting := "Welcome to the Tic-Tac-Toe Server."
	_, err := conn.Write([]byte(greeting))
	if err != nil {
		fmt.Println("[Server] Error sending greeting message:", err)
		return
	}

	buf := make([]byte, 2048)
	for {

		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("[Server|%s] Error reading from client: %s\n", conn.RemoteAddr().String(), err)
			return
		}
		payload := buf[:n]
		s.ch <- Message{
			sender:  conn,
			payload: payload,
		}

	}

}

