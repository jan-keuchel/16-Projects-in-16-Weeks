package main

import (
	"fmt"
	"net"
	"os"
)

type Client struct {
	serverAddr string
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) connectToServer() {

	fmt.Println("Connecting to Server at", c.serverAddr, "...")
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		fmt.Println("[Client] Error dialing server. Exiting...")
		return
	}
	fmt.Println("[Client] Successfully connected to server.")

	go c.listenToServer(conn)

	//TODO: Take input from terminal

}

func (c *Client) listenToServer(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("[Client] Error reading message from server:", err)
			continue
		}
		data := buf[:n]
		fmt.Println("Received:\n", data)
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
		
		// TODO: Start Server
		

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
