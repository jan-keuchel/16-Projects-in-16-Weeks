package main

import (
	"fmt"
	"os"
)

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
		// TODO: Start Server
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
	} else if typeInput == "C" {
		// TODO: Start Client
		fmt.Println("Enter IP-address and port to connect to (<IP>:<Port>):")
		var serverAddr string
		_, err := fmt.Scan(&serverAddr)
		if err != nil {
			fmt.Println("Error reading input (port):", err)
			os.Exit(3)
		}
		fmt.Println("Connecting to Server at", serverAddr, "...")
	} else {
		fmt.Println("Invalid input (", typeInput, "). Use 'S' or 'C'.")
		os.Exit(2)
	}

}
