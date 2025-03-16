package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Message struct {
	sender  net.Conn
	payload []byte
}

type Server struct {
	listenAddr 	 string
	clientConns  map[net.Conn] bool
	msgChannel	 chan Message
	mu  		 sync.Mutex
}

func NewServer(listenAddr string) *Server {

	return &Server{
		listenAddr:   listenAddr,
		clientConns:  make(map[net.Conn]bool),
		msgChannel:   make(chan Message),
	}

}

func (s *Server) Start() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(2)
	go s.processMessageChannelInput(ctx, &wg)
	go s.acceptClientConnections(ctx, &wg)

	select {
	case <-sigCh:
		fmt.Println("[Log] Server Received shutdown signal. Initiating shutdown...")
		cancel()
	}

	wg.Wait()

	fmt.Println("[Log] Shutdown complete.")

}

// acceptClientConnections starts a listener on the servers specified address and
// accepts incomming client connections. The accepted client connection is then
// handled by handleClientConnection in a separate goroutine.
//
// The following operations are performed:
// 	 - Starting up the listener
// 	 - Waiting for and accepting incomming connections
// 	 - Starting a new goroutine per incomming connection
func (s *Server) acceptClientConnections(ctx context.Context, mainWG *sync.WaitGroup) {

	defer mainWG.Done()

	var wg sync.WaitGroup

	fmt.Println("[Log] Setting up listener...")
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		fmt.Println("[Error] Starting to listen for connections:", err)
		return
	}
	defer ln.Close()

	fmt.Println("[Log] Now listening for incomming client connections...")
	for {

		if err := ln.(*net.TCPListener).SetDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			fmt.Println("[Log] Error setting deadline for incomming client connections:", err)
			return
		}

		select {
		case <-ctx.Done():
			fmt.Println("[Log] Shutting down listener...")
			fmt.Println("[Log] Waiting for client handlers to termiante...")
			wg.Wait()
			fmt.Println("[Log] All client handlers terminated.")
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Println("[Error] Accepting incomming connection:", err)
				continue
			}

			wg.Add(1)
			go s.handleClientConnection(ctx, &wg, conn)

		}
	}

}

// handleClientConnection manages a single client connection for the server.
// It handles reading messages from the client, forwarding them to the server's
// message channel, and properly cleaning up resources when the connection is closed.
//
// The function runs in its own goroutine for each client connection and performs
// the following key operations:
//   - Registers the client connection with the server
//   - Creates a quit channel for proper shutdown
//   - Reads incoming messages into a 2048-byte buffer
//   - Forwards messages to the server's message channel
//   - Handles connection errors and EOF conditions
//   - Ensures proper cleanup through defer statements
//
// Parameters:
//   conn - The net.Conn representing the client connection
//
// All connection state changes are protected by the server's mutex to ensure
// thread safety. Connection closure is logged with the client's remote address.
func (s *Server) handleClientConnection(ctx context.Context, 
	 									wg *sync.WaitGroup, 
	 									conn net.Conn) {

	defer func() {
		s.mu.Lock()
		delete(s.clientConns, conn)
		s.mu.Unlock()
		conn.Close()
		wg.Done()
	}()

	s.mu.Lock()
	s.clientConns[conn] = true
	s.mu.Unlock()

	b := make([]byte, 2048)
	for {

		err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if err != nil {
			fmt.Printf("[Error|%s] Setting up Read deadline:\n%s\n",conn.RemoteAddr(),  err)
			return
		}

		select {
		case <-ctx.Done():
			fmt.Println("[Log] Shutting down client handler...")
			return
		default:
			n, err := conn.Read(b)
			if err != nil {
				// checks if err implements the net.Error interface
				// if it does (ok = true), checks if a Timeout occured
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if err == io.EOF {
					fmt.Printf("[Log|%s] Client closed connection.\n", conn.RemoteAddr())
					return
				}
				fmt.Println("[Error] Reading message from client:", err)
				return
			}
			payload := b[:n]
			s.msgChannel <- Message{
				sender:  conn,
				payload: payload,
			}
		}

	}

}

func (s *Server) processMessageChannelInput(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {
		case <-ctx.Done():
			fmt.Println("[Log] Shutting down processing of input...")
			return
		case msg :=<- s.msgChannel:
			fmt.Printf("[Log] Message received from %s:\n%s\n", msg.sender.RemoteAddr(), string(msg.payload))

			pld := string(msg.payload)
			if strings.HasPrefix(pld, "/") {
				fmt.Printf("[Log] Message from %s is a command: '%s'.", msg.sender.RemoteAddr(), pld)

			}
		}

	}

}


// -----------------------------
// ---------- Handler ----------
// -----------------------------

