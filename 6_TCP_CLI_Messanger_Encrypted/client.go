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
	quitCh 	   chan bool
}

func NewClient(serverAddr string) *Client {
	
	return &Client{
		serverAddr: serverAddr,
		quitCh:     make(chan bool),
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
	defer func() {
		_, ok := <- c.quitCh
		if ok {
			close(c.quitCh)
		}
		conn.Close()
	}()

	fmt.Println("[Log] Connection established.")

	go c.listenToServer(conn)

	inputCh := make(chan string)
	errCh   := make(chan error)

	go func() {

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputCh <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
		} else {
			errCh <- io.EOF
		}

	}()

	for {

		select {
		case <-c.quitCh:
			fmt.Println("[Log] Stopping client due to server disconnection.")
			return
		case input := <- inputCh:
			_, err := conn.Write([]byte(input))
			if err != nil {
				fmt.Println("[Error] Writing to server:", err)
				return
			}
		case err := <- errCh:
			if err != io.EOF {
				fmt.Println("[Error] Reading input from stdin:", err)
			} else {
				fmt.Println("[Error] Input ended (EOF)")
			}
			return
		}

	}

}

func (c *Client) listenToServer(conn net.Conn) {

	b := make([]byte, 2048)
	for {

		select {
		case <- c.quitCh:
			fmt.Printf("[Log] No longer listening to server.")
			return
		default:
			n, err := conn.Read(b)
			if err != nil {
				if err == io.EOF {
					fmt.Println("[Log] Server closed connection.")
					close(c.quitCh)
					return
				}
				fmt.Println("[Error] Reading from server:", err)
				return
			}

			payload := b[:n]

			fmt.Printf("Message from server:\n%s\n", string(payload))
		}

	}

}

