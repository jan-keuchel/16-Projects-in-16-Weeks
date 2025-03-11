package main

import (
	"fmt"
	"io"
	"net"
)

type Client struct {
	serverAddr string
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
		fmt.Println("[Client] Error dialing server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("[Client] Successfully connected to server.")

	// TODO: properly terminate Goroutine if server is full.
	go c.listenToServer(conn)

	var input string
	for {

		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println("[Client] Error reading input:", err)
			return
		}
		// fmt.Println("[Client] Read from terminal:", input)

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
			if err == io.EOF {
				fmt.Println("[Client] Server closed connection.")
				return
			}
			fmt.Println("[Client] Error reading message from server:", err)
			return
		}
		data := buf[:n]
		fmt.Printf("[Cleint] Received:\n%s\n", string(data))
	}

}

