package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	sender  net.Conn
	payload []byte
}

type Server struct {
	listenAddr 	 string
	clientConns  map[net.Conn]string
	ch  		 chan Message
	mu 			 sync.Mutex
	game 		 *TTT
	activeClient net.Conn
}

// Returns a pointer to an initialized Server
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:  listenAddr,
		ch:  		 make(chan Message),
		clientConns: make(map[net.Conn]string),
		game:        nil,
	}
}

// Sends 'msg' to every client of the server.
func (s *Server) broadcast(msg string) {

	s.mu.Lock()
	for clientConn := range s.clientConns {
		n, err := clientConn.Write([]byte(msg))
		if err != nil {
			fmt.Println("[Server] Error while broadcasting to", clientConn.RemoteAddr(), ":", err)
			continue
		}
		if n != len(msg) {
			fmt.Println("[Server] Error: Couldn't send entire message at broadcast to", clientConn.RemoteAddr(), ".")
			continue
		}
	}
	s.mu.Unlock()

}

// sendActivePlayerMessage broadcasts the information of whos
// turn it is to the players.
func (s *Server) sendActivePlayerMessage() {

	s.mu.Lock()
	for client := range s.clientConns {

		if client == s.activeClient {
			_, err := client.Write([]byte("You are the active player. Make your move:"))
			if err != nil {
				fmt.Println("Error writing active player message to active player:", err)
				continue
			}
		} else {
			_, err := client.Write([]byte("It's not your turn. Waiting..."))
			if err != nil {
				fmt.Println("Error writing active player message to inactive player:", err)
				continue
			}
		}

	}
	s.mu.Unlock()

}

// startGame initializes the game, sets up the figures, chooses starting player,
// broadcasts the game board and sends the active player message.
func (s *Server) startGame() {

	s.game = NewTTT()

	s.mu.Lock()
	setX := false
	for conn := range s.clientConns {
		if !setX {
			s.clientConns[conn] = "X"
			s.activeClient = conn
			setX = true
		} else {
			s.clientConns[conn] = "O"
		}
	}
	s.mu.Unlock()

	s.broadcast(s.game.printBoard())
	s.sendActivePlayerMessage()

}

// Receives data sent to the server via channel, calls function to process data and
// sends back the response for the clients
func (s *Server) processClientInput() {

	for msg := range s.ch {

		fmt.Printf("[%s]: %s\n", msg.sender.RemoteAddr(), string(msg.payload))

		if len(s.clientConns) == 1 {
			s.broadcast("You're all alone. Why are you talking?")
			continue
		}

		if msg.sender != s.activeClient {
			_, err := msg.sender.Write([]byte("It's not your turn. Please wait!"))
			if err != nil {
				fmt.Println("[Server] Error writing 'not your turn' message:", err)
				continue
			}
			// TODO: Handle disconnect request
		} else {

			if string(msg.payload) < "0" || string(msg.payload) > "8" {
				fmt.Println("[Server] Client didn't sent num in [0;8].")
				_, err := msg.sender.Write([]byte("Invalid input: Please send a number from 0 to 8."))
				if err != nil {
					fmt.Println("[Server] Error writing wrong input message:", err)
					continue
				}
			}

			cell, err := strconv.Atoi(string(msg.payload))
			if err != nil {
				fmt.Println("[Server] Error converting input to int.")
				continue
			}

			validMove, tiedGame := s.game.stepGame(cell, s.clientConns[s.activeClient])
			if !validMove {
				fmt.Println("[Server] Client sent taken cell number.")
				_, err := msg.sender.Write([]byte("That cell is already taken."))
				if err != nil {
					fmt.Println("[Server] Error writing taken cell message:", err)
					continue
				}
			} else {
				if tiedGame {
					fmt.Println("[Server] Tied game!")
					_, err := msg.sender.Write([]byte("It seems like we have a tie here..."))
					if err != nil {
						fmt.Println("[Server] Error writing tied game message:", err)
						s.broadcast(s.game.printBoard())
						continue
					}
				} else if s.game.checkForWinner(cell, s.clientConns[s.activeClient]) {
					winMsg := s.clientConns[s.activeClient] + " won the game."
					s.broadcast(winMsg)
				} else {
					s.switchActiveConnection()
				}
			}

			s.broadcast(s.game.printBoard())
			s.sendActivePlayerMessage()

		}

	}

}

// Sets up a listener on specified port and waits for clients to connect. 
// Only accepts incomming connection if there are less than 2 clients connected.
func (s *Server) acceptClients() {

	fmt.Println("[Server] Setting up listener...")
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		fmt.Println("[Server] Error listening for client connections:", err)
		return
	}
	defer ln.Close()
	fmt.Println("[Server] Waiting for clients to connect...")

	for {

		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[Server] Error accepting incomming client connection:", err)
			continue
		}

		fmt.Println("[Server] New Connection request from client.")

		if len(s.clientConns) < 2 {
			s.clientSetup(conn)
			go s.listenToClientConnection(conn)
		} else {
			fmt.Println("[Server] Connection request discarded: Already 2 clients.")
			_, err := conn.Write([]byte("Server is already full. Closing connection..."))
			if err != nil {
				fmt.Println("[Server] Error writing rejection due to full server:", err)
				continue
			}
			conn.Close()
		}

	}

}

// clientSetup manages the client connections of the server and sends a greeting to
// the client
func (s *Server) clientSetup(conn net.Conn) {

	fmt.Println("[Server] Adding client to server...")

	s.mu.Lock()
	s.clientConns[conn] = "T" // Temporary
	s.mu.Unlock()

	fmt.Println("[Server] New Client:", conn.RemoteAddr())

	_, err := conn.Write([]byte("Welcome to the Tic-Tac-Toe Server."))
	if err != nil {
		fmt.Println("[Server] Error sending greeting message:", err)
		return
	}

	if len(s.clientConns) == 1 {

		_, err := conn.Write([]byte("You are currently the only client. Waiting for another player to join..."))
		if err != nil {
			fmt.Println("[Server] Error sending greeting message:", err)
			return
		}

	} else if len(s.clientConns) == 2 {

		s.broadcast("2 Players connected. Starting game...")
		time.Sleep(time.Second)

		s.startGame()

	}


}

// updateActiveConnection sets the servers activeClient - meaning the client who
// would make the next move - to the given connection if that connection is still
// present.
func (s *Server) updateActiveConnection(conn net.Conn) {

	s.mu.Lock()
	if s.clientConns[conn] != "" {
		s.activeClient = conn
	}
	s.mu.Unlock()

}

// switchActiveConnection changes the active player to the player who is not currently
// the active player.
func (s *Server) switchActiveConnection() {

	s.mu.Lock()
	for client := range s.clientConns {
		if client != s.activeClient {
			s.activeClient = client
			break
		}
	}
	s.mu.Unlock()

}

// Indefinitely reads from TCP connection and sends the read data 
// into the servers channel.
func (s *Server) listenToClientConnection(conn net.Conn) {

	defer conn.Close()
	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		s.mu.Unlock()
	}()
	
	buf := make([]byte, 2048)
	for {

		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("[Server|%s] Error reading from client: %s\n", conn.RemoteAddr().String(), err)
			return
		}
		payload := buf[:n]
		s.ch <- Message{
			sender:  conn,
			payload: payload,
		}

	}

}

