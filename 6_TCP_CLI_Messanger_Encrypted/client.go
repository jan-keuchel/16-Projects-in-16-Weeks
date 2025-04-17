package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type CommandPreprocesser func(c *Client, conn net.Conn, payload string) (string, error)

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
	serverAddr     string
	quitCh 	       chan bool
	inDialogueMode bool
}

func NewClient(serverAddr string) *Client {
	
	return &Client{
		serverAddr:     serverAddr,
		quitCh:         make(chan bool),
		inDialogueMode: false,
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

		reader := bufio.NewReader(os.Stdin)
		for  {

			if c.inDialogueMode {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			fmt.Print("> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("[Error] Reading input from Stdin:", err)
				close(inputCh)
				return
			}
			inputCh <- strings.TrimSpace(input)

		}

	}()

	for {

		select {
		case <-c.quitCh:
			fmt.Println("[Log] Stopping client due to server disconnection.")
			return
		case input := <- inputCh:
			msgType := "MESSAGE"

			if strings.HasPrefix(input, "/") {

				command := strings.Fields(input)[0]
				preproFunc, ok := commandRequirementFunctions[command]
				if !ok {
					fmt.Printf("[Error] Command is not valid: %s\n", command)
					continue
				}

				preprocessedCommand, err := preproFunc(c,conn, input)
				if err != nil {
					fmt.Println("[Error] Wrong use of command:", err)
					continue
				}

				input   = preprocessedCommand
				msgType = "COMMAND"

			}

			errMsg := "Writing message to server failed"
			c.sendMessageToServer(conn, input, msgType, errMsg)
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

func (c *Client) sendMessageToServer(conn net.Conn, payload string, msgType string, errMsg string) {

	packet := Packet{
		MsgType: msgType,
		Payload: payload,
	}

	jsonData, errJ := json.Marshal(packet)
	if errJ != nil {
		fmt.Println("[Error] Marshalling message to json format:", errJ)
		return
	}

	length := uint32(len(jsonData))
	errLen := binary.Write(conn, binary.BigEndian, length)
	if errLen != nil {
		fmt.Println("[Error] Sending message length:", errLen)
		return
	}

	_, err := conn.Write(jsonData)
	if err != nil {
		fmt.Println(errMsg + ":", err)
	}

}

func (c *Client) listenToServer(conn net.Conn) {

	for {

		select {
		case <- c.quitCh:
			fmt.Printf("[Log] No longer listening to server.")
			return
		default:

			var length uint32
			errLen := binary.Read(conn, binary.BigEndian, &length)
			if errLen != nil {
				if errLen == io.EOF || errLen == io.ErrUnexpectedEOF {
					fmt.Println("[Log] Server closed connection.")
					close(c.quitCh)
					return
				}
				fmt.Println("[Error] Reading length from server:", errLen)
				close(c.quitCh)
				return
			}
			data := make([]byte, length)

			_, err := io.ReadFull(conn, data)
			if err != nil {
				if err == io.EOF {
					fmt.Println("[Log] Server closed connection.")
					close(c.quitCh)
					return
				}
				fmt.Println("[Error] Reading from server:", err)
				return
			}

			var packet Packet
			errUm := json.Unmarshal(data, &packet)
			if errUm != nil {
				fmt.Println("[Error] Unmarshalling server data failed:", errUm)
				return
			}

			fmt.Printf("Received data from server.\nType: %s\nPayload:%s\n", packet.MsgType, string(packet.Payload))
		}

	}

}

// -----------------------------
// ---------- Handler ----------
// -----------------------------

func preprocessQuit(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(payload)) != 1 {
		return "", errors.New("'/quit' command was given the wrong number of arguments. Please just use '/quit' without any further arguments in order to quit the connection to the server.")
	}
	return "/quit", nil

}

func preprocessRegister(c *Client, conn net.Conn, payload string) (string, error) {

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

func preprocessHelp(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(payload)) != 1 {
		return "", errors.New("'/help' command was given the wrong number of arguments. Plese just use '/help' without any further arguments in order to show a list of available commands and their usecases.")
	}
	return "/help", nil

}

func preprocessLogin(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/login' command was given the wrong number of arguments. Please just use '/login' without any other arguments in order to start the login process.")
	}

	c.inDialogueMode = true
	defer func ()  {
		c.inDialogueMode = false
	}();

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Login-process started. Enter your username and press 'Return':")
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Println("[Error] Reading from stdin:", err)
		}		
		return "", errors.New("'/login' command failed to read username.")
	}
	username := scanner.Text()

	fmt.Println("Now enter your password (input hidden):")
	pwd, errPW := term.ReadPassword(int(os.Stdin.Fd()))
	if errPW != nil {
		fmt.Println("[Error] Reading password failed:", errPW)
		return "", errPW
	}
	fmt.Println()

	hash := sha256.New()
	_, err := hash.Write(pwd)
	if err != nil {
		return "", err
	}
	pwdHsh := hash.Sum(nil)

	res := payload + " " + username + " " + string(pwdHsh)
	fmt.Println("[Log] Now sending login data to the server...")

	return res, nil

}

func preprocessLogout(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/logout' command was given the wrong number of arguments. Plese just use '/logout' without any further arguments in order to log out of the user account you're currently logged in as.")
	}
	return "/logout", nil

}

func preprocessNewChat(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 2 {
		return "", errors.New("'/newChat' command was given the wrong number of arguments. Plese use '/newChat <username>' in order to send a request to <username>.")
	}

	slicedPld := strings.Fields(payload)
	command   := slicedPld[0]
	user      := slicedPld[1]

	res := command + " " + user 

	return res, nil

}

func preprocessAccept(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/accept' command was given the wrong number of arguments. Plese just use '/accept' without any further arguments in order to accept the current chat request..")
	}
	return "/accept", nil


}

func preprocessDecline(c *Client, conn net.Conn, payload string) (string, error) {

	if len(strings.Fields(string(payload))) != 1 {
		return "", errors.New("'/decline' command was given the wrong number of arguments. Plese just use '/decline' without any further arguments in order to decline the current chat request..")
	}
	return "/decline", nil

}
