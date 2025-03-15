package main

import (
	"bufio"
	"fmt"
	"os"
)

func startServer() {

	fmt.Println("Start server")

}

func startClient() {

	fmt.Println("Start client")

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
				fmt.Println("Error reading from stdin:", err)
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
		startServer()
	case "c", "C":
		startClient()
	}

}
