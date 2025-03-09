package main

import (
	"fmt"
	"net"
	"sync"
	"time"
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

// Sends 'msg' to every client of the server.
func (s *Server) broadcast(msg string) {

	s.mu.Lock()
	for clientConn := range s.clientConns {
		n, err := clientConn.Write([]byte(msg))
		if err != nil {
			fmt.Println("[Server] Error while broadcasting to", clientConn.RemoteAddr(), ":", err)
			continue
		}
		if n != len(msg) {
			fmt.Println("[Server] Error: Couldn't send entire message at broadcast to", clientConn.RemoteAddr(), ".")
			continue
		}
	}
	s.mu.Unlock()

}

// Receives data sent to the server via channel, calls function to process data and
// sends back the response for the clients
func (s *Server) processClientInput() {

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

	_, err := conn.Write([]byte("Welcome to the Tic-Tac-Toe Server."))
	if err != nil {
		fmt.Println("[Server] Error sending greeting message:", err)
		return
	}
	if len(s.clientConns) == 1 {
		_, err := conn.Write([]byte("You are currently the only client. Waiting for another player to join."))
		if err != nil {
			fmt.Println("[Server] Error sending greeting message:", err)
			return
		}
	} else if len(s.clientConns) == 2 {
		s.broadcast("2 Players connected. Starting game...")
		time.Sleep(time.Second)
		// TODO: Start game
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

