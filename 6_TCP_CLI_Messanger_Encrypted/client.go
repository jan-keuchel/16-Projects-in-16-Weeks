package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type CommandPreprocesser func(c *Client, payload string) (string, error)

var commandRequirementFunctions = map[string]CommandPreprocesser {
	"/quit": 		preprocessQuit,
	"/register": 	preprocessRegister,
	"/help": 		preprocessHelp,
	"/login": 		preprocessLogin,
	"/logout": 		preprocessLogout,
	"/newChat": 	preprocessNewChat,
	"/accept": 		preprocessAccept,
	"/decline": 	preprocessDecline,
}

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
		}

	}()

	for {

		select {
		case <-c.quitCh:
			fmt.Println("[Log] Stopping client due to server disconnection.")
			return
		case input := <- inputCh:
			if strings.HasPrefix(input, "/") {

				command := strings.Fields(input)[0]
				preproFunc, ok := commandRequirementFunctions[command]
				if !ok {
					fmt.Printf("[Error] Command is not valid: %s\n", command)
					continue
				}

				preprocessedCommand, err := preproFunc(c, input)
				if err != nil {
					fmt.Println("[Error] Wrong use of command:", err)
					continue
				}

				input = preprocessedCommand

			}
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

// -----------------------------
// ---------- Handler ----------
// -----------------------------

func preprocessQuit(c *Client, payload string) (string, error) {

	if len(strings.Fields(payload)) != 1 {
		return "", errors.New("'/quit' command was given the wrong number of arguments. Please just use '/quit' without any further arguments in order to quit the connection to the server.")
	}
	return "/quit", nil

}

func preprocessRegister(c *Client, payload string) (string, error) {

	// TODO: Add character check. Only alphabetic characters for username permitted.

	if len(strings.Fields(payload)) != 3 {
		return "", errors.New("'/register' command was given the wrong number of arguments. Please provide username and password according to the following pattern: '/register <username> <password>'.")
	}

	slicedPld := strings.Fields(payload)
	command   := slicedPld[0]
	username  := slicedPld[1]
	pwd   	  := []byte(slicedPld[2])

	hash := sha256.New()
	_, err := hash.Write(pwd)
	if err != nil {
		return "", err
	}
	pwdHsh := hash.Sum(nil)

	res := command + " " + username + " " + string(pwdHsh)

	return res, nil

}

func preprocessHelp(c *Client, payload string) (string, error) {

	if len(strings.Fields(payload)) != 1 {
		return "", errors.New("'/help' command was given the wrong number of arguments. Plese just use '/help' without any further arguments in order to show a list of available commands and their usecases.")
	}
	return "/help", nil

}

func preprocessLogin(c *Client, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 3 {
		return "", errors.New("'/login' command was given the wrong number of arguments. Please provide username and password according to the following pattern: '/login <username> <password>'.")
	}

	slicedPld := strings.Fields(payload)
	command   := slicedPld[0]
	username  := slicedPld[1]
	pwd   	  := []byte(slicedPld[2])

	hash := sha256.New()
	_, err := hash.Write(pwd)
	if err != nil {
		return "", err
	}
	pwdHsh := hash.Sum(nil)

	res := command + " " + username + " " + string(pwdHsh)

	return res, nil

}

func preprocessLogout(c *Client, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/logout' command was given the wrong number of arguments. Plese just use '/logout' without any further arguments in order to log out of the user account you're currently logged in as.")
	}
	return "/logout", nil

}

func preprocessNewChat(c *Client, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 2 {
		return "", errors.New("'/newChat' command was given the wrong number of arguments. Plese use '/newChat <username>' in order to send a request to <username>.")
	}

	slicedPld := strings.Fields(payload)
	command   := slicedPld[0]
	user      := slicedPld[1]

	res := command + " " + user 

	return res, nil

}

func preprocessAccept(c *Client, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/accept' command was given the wrong number of arguments. Plese just use '/accept' without any further arguments in order to accept the current chat request..")
	}
	return "/accept", nil


}

func preprocessDecline(c *Client, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/decline' command was given the wrong number of arguments. Plese just use '/decline' without any further arguments in order to decline the current chat request..")
	}
	return "/decline", nil

}
