package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode"
)

var server Server

func isNumeric(s string) bool {

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true

}

func startServer(port string) {

	fmt.Println("Starting server...")
	listenAddr := ":" + port
	server := NewServer(listenAddr)
	server.Start()

}

func startClient(serverAddr string) {

	fmt.Println("Starting client...")
	client := NewClient(serverAddr)
	client.connectToServer()

}

func main() {

	fmt.Println("CLI E2EE Messanger")
	fmt.Println("This applicatoin can either be started as a server or a client. Choose by typing either 's' for Server or 'c' for client:")
	scanner := bufio.NewScanner(os.Stdin)

	// Receive input to either start server or client
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
		if input == "s" || input == "S" || input == "c" || input == "C" {
			break
		}

		fmt.Println("Received wrong input. Please enter an 's' if you wish to start the server or a 'c' if you wish to start as a client:")

	}

	switch input {
	case "s", "S":
		fmt.Println("Server setup:\nEnter the port to listen on for incomming connections:")
		var port string
		for {

			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					fmt.Println("Error reading from stdin:", err)
				} else {
					fmt.Println("Input ended. (EOF)")
				}
				return
			}

			port = scanner.Text()
			if isNumeric(port) {
				break
			}

			fmt.Println("Received wrong input. Please enter a number to specify the port the server will listen on for incomming connections:")

		}

		startServer(port)
		
	case "c", "C":
		fmt.Println("Client setup:\nEnter the IP address of the server to connect to:")
		var ip string
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading from stdin:", err)
			} else {
				fmt.Println("Input ended. (EOF)")
			}
			return
		}
		ip = scanner.Text()
		// TODO: Verify IP address
		
		fmt.Printf("Client setup:\nNow enter the port the server at '%s' is listening on:\n", ip)
		var port string
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading from stdin:", err)
			} else {
				fmt.Println("Input ended. (EOF)")
			}
			return
		}
		port = scanner.Text()
		// TODO: Verify port

		serverAddr := ip + ":" + port
		startClient(serverAddr)

	}

}
