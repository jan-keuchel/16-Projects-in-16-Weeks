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

		fmt.Println("Enter the port to listen on:")
		var port string
		_, err := fmt.Scan(&port)
		if err != nil {
			fmt.Println("Error reading input (port):", err)
			os.Exit(3)
		}
		listenAddr := ":" + port
		fmt.Println("Starting server on port", port, "...")
		
		server := NewServer(listenAddr)
		go server.processClientInput()
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
