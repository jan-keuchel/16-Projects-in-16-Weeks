package main

import (
	"bufio"
	"fmt"
	"os"
)

var server Server


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
		fmt.Println("\nServer setup:\nEnter the port to listen on for incomming connections:")
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
		fmt.Println("\nClient setup:\nEnter the IP address of the server to connect to:")
		var ip string
		for {

			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					fmt.Println("Error reading from stdin:", err)
				} else {
					fmt.Println("Input ended. (EOF)")
				}
				return
			}
			ip = scanner.Text()

			if verifyIPFormat(ip) {
				break
			}

			fmt.Println("Received wrong input. Please provide a valid IP address. That means: 4 numbers in the range of 0 - 255 separated by '.':")

		}

		fmt.Printf("\nClient setup:\nNow enter the port the server at '%s' is listening on:\n", ip)
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

			fmt.Printf("Received wrong input. Please enter a number to specify the port the server at '%s' is listening on:\n", port)

		}

		serverAddr := ip + ":" + port
		startClient(serverAddr)

	}

}
