package main

import (
	"bufio"
	"fmt"
	"io"
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

// connectToServer establishes a connection to the server specified
// by the client's serverAddr field. A goroutine which is listening to
// messages from the server is launched and the continual input read 
// from stdin is sent to the server.
func (c *Client) connectToServer() {

	fmt.Println("[Log] Dialing server...")
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		fmt.Println("[Error] Dialing server at", c.serverAddr, ":", err)
		return
	}
	defer conn.Close()
	fmt.Println("[Log] Connection established.")

	go c.listenToServer(conn)

	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for {

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("[Error] Reading from stdin:", err)
			} else {
				fmt.Println("Input ended. (EOF)")
			}
			return
		}

		input = scanner.Text()

		_, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Println("[Error] Writing to server:", err)

		}

	}

}

func (c *Client) listenToServer(conn net.Conn) {

	b := make([]byte, 2048)
	for {

		n, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				fmt.Println("[Log] Server closed connection.")
				return
			}
			fmt.Println("[Error] Reading from server:", err)
			return
		}

		payload := b[:n]

		fmt.Printf("Message from server:\n%s\n", string(payload))

	}

}

