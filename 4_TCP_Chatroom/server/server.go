package main

import (
	"fmt"
	"net"
	"sync"
)

// --------------- TASK 4 ---------------

type Message struct {
	senderConn  net.Conn
	payload  	[]byte
}

type Server struct {
	clientConns map[net.Conn]bool
	ch 	 	 	chan Message
	mu 			sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clientConns: make(map[net.Conn]bool),
		ch:  	  	 make(chan Message, 20),
	}
}

func (s *Server) handleConnection(conn net.Conn) {

	defer conn.Close()

	s.mu.Lock()
	s.clientConns[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		s.mu.Unlock()
	}()

	addr := conn.RemoteAddr().String()

	for {

		fmt.Println("Server waiting for data from", addr, "...")
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			fmt.Println("Error while reading from", addr, ":", err)
			return
		}

		data := b[:n]
		fmt.Println("Server read data from", addr, "successfully:", string(data))

		fmt.Println("Sending data of", addr, "to broadcasting routine...")
		s.ch <- Message{
			senderConn:  conn,
			payload:  	 data,
		}

	}


}

func (s *Server) startListeningForclientConns() {

	fmt.Println("Server is listening...")
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Server error while listening:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server listening on Port 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Server error at connection Accept:", err)
			continue
		}
		fmt.Println("Server accepted client connection.")
		go s.handleConnection(conn)
	}


}

func (s *Server) broadcast() {

	for msg := range s.ch {

		senderString := "[" + msg.senderConn.RemoteAddr().String() + "]: "
		data := append([]byte(senderString), msg.payload...)

		s.mu.Lock()
		for client := range s.clientConns {
			_, err := client.Write(data)
			if err != nil {
				fmt.Println("Server error at writing to", client.RemoteAddr().String(), ":", err)
				delete(s.clientConns, client)
				client.Close()
				continue
			}
		}
		s.mu.Unlock()

	}

}

func task4() {

	server := NewServer()
	go server.broadcast()
	server.startListeningForclientConns()

}

func main() {

	task4()

}
