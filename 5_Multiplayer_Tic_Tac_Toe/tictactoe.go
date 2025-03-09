package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type Message struct {
	sender  net.Conn
	payload []byte
}

type Client struct {
	serverAddr string
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

		if len(s.clientConns) < 2 {
			go s.listenToClientConnection(conn)
		}

	}

}

// Adds connection (client) to the map of clients of the server, indefinitely
// reads from TCP connection and sends the read data into the servers channel.
func (s *Server) listenToClientConnection(conn net.Conn) {

	defer conn.Close()
	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		s.mu.Unlock()
	}()

	s.mu.Lock()
	s.clientConns[conn] = true
	s.mu.Unlock()

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

// Returns a pointer to a initialized Client struct
func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

// Establishes a connection to the server with the address specified in the Client
// struct. Indefinitely reads input from stdin and writes read data to server.
// Launches Goroutine to listen for server data.
func (c *Client) connectToServer() {

	fmt.Println("[Client] Connecting to Server at", c.serverAddr, "...")
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		fmt.Println("[Client] Error dialing server. Exiting...")
		return
	}
	defer conn.Close()
	fmt.Println("[Client] Successfully connected to server.")

	go c.listenToServer(conn)

	var input string
	for {

		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println("[Client] Error reading input:", err)
			return
		}
		fmt.Println("[Client] Read from terminal:", input)

		n, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Println("[Client] Error writing to server:", err)
			continue
		}
		if n != len(input) {
			fmt.Printf("[Client] Error: Couldn't write entire input\n")
			continue
		}

	}

}

// Indefinitely reads messages via 'conn' from server and prints to stdout.
func (c *Client) listenToServer(conn net.Conn) {

	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("[Client] Error reading message from server:", err)
			continue
		}
		data := buf[:n]
		fmt.Println("[Cleint] Received:\n", data)
	}

}

func main() {

	fmt.Println("Welcome to Tic-Tac-Toe!")
	fmt.Println("[S]erver or [C]lient: ")

	var typeInput string
	_, err := fmt.Scan(&typeInput)
	if err != nil {
		fmt.Println("Error reading input (Server/Client):", err)
		os.Exit(1)
	}

	if typeInput == "S" {

		fmt.Println("Enter the port to listen on:")
		var listenAddr string = "localhost:"
		var port string
		_, err := fmt.Scan(&port)
		if err != nil {
			fmt.Println("Error reading input (port):", err)
			os.Exit(3)
		}
		listenAddr += port
		fmt.Println("Starting server on port", port, "...")
		
		server := NewServer(listenAddr)
		go server.broadcast()
		server.acceptClients()

	} else if typeInput == "C" {

		fmt.Println("Enter IP-address and port to connect to (<IP>:<Port>):")
		var serverAddr string
		_, err := fmt.Scan(&serverAddr)
		if err != nil {
			fmt.Println("Error reading input (port):", err)
			os.Exit(3)
		}

		client := NewClient(serverAddr)
		client.connectToServer()

	} else {

		fmt.Println("Invalid input (", typeInput, "). Use 'S' or 'C'.")

		os.Exit(2)
	}

}
