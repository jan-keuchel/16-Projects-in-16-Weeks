package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func listenForMessages(conn net.Conn) {

	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			fmt.Println("Client error while reading:", err)
			continue
		}

		msg := b[:n]
		fmt.Printf("Received: %s", string(msg))
	}

}

func main() {

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Client error while dialing server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Successfully connected to server.")

	go listenForMessages(conn)

	scanner := bufio.NewScanner(os.Stdin) 
	for {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Client error reading stdin:", err)
			} else {
				fmt.Println("Input ended (EOF)")
			}
		}
		line := scanner.Text() + "\n"

		n, err := conn.Write([]byte(line))
		if err != nil {
			fmt.Println("Client error while writing:", err)
			continue
		}
		if n != len(line) {
			fmt.Println("Client error: Couldn't write entire line.")
			continue
		}
	}

}
