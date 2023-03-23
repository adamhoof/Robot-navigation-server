package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_ADDRESS = "localhost"
	SERVER_PORT    = "3999"
	PROTOCOL       = "tcp"
)

const (
	TERMINATOR = "\a\b"

	MAX_NAME_LEN  = 18
	MIN_KEY_INDEX = 0
	MAX_KEY_INDEX = 4
)

type ServerMessage string

const (
	UNABLE_TO_CLOSE_SOCKET ServerMessage = "unable to cutOff conn\n"

	SERVER_KEY_REQUEST ServerMessage = "107 KEY REQUEST" + TERMINATOR

	SERVER_MOVE ServerMessage = "102 MOVE" + TERMINATOR

	SERVER_TURN_LEFT  ServerMessage = "103 TURN LEFT" + TERMINATOR
	SERVER_TURN_RIGHT ServerMessage = "104 TURN RIGHT" + TERMINATOR
	SERVER_PICK_UP    ServerMessage = "105 GET MESSAGE" + TERMINATOR
	SERVER_LOGOUT     ServerMessage = "106 TURN LEFT" + TERMINATOR

	SERVER_OK                     ServerMessage = "200 OK" + TERMINATOR
	SERVER_LOGIN_FAILED           ServerMessage = "300 LOGIN FAILED" + TERMINATOR
	SERVER_SYNTAX_ERROR           ServerMessage = "301 SYNTAX ERROR" + TERMINATOR
	SERVER_LOGIC_ERROR            ServerMessage = "302 LOGIC ERROR" + TERMINATOR
	SERVER_KEY_OUT_OF_RANGE_ERROR ServerMessage = "303 KEY OUT OF RANGE" + TERMINATOR
)

type State int

const (
	USERNAME   State = 0
	KEY        State = 1
	VALIDATION State = 2
	MOVE       State = 3
)

type MoveState int

const (
	DERIVE_POS MoveState = 0
	STRAIGHT   MoveState = 1
	RIGHT      MoveState = 2
	LEFT       MoveState = 3
)

type KeyPair struct {
	ServerKey int
	ClientKey int
}

type Client struct {
	conn  *net.Conn
	Name  string
	KeyID int
	Hash  int
	pos   Position
}

type Position struct {
	x int
	y int
}

func createListener() (net.Listener, error) {
	fmt.Printf("Starting server on %s:%s\n", SERVER_ADDRESS, SERVER_PORT)
	listener, err := net.Listen(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if err != nil {
		return listener, err
	}
	fmt.Println("Server started...")
	return listener, err
}

func waitForClientConnection(listener *net.Listener) (net.Conn, error) {
	log.Println("Waiting for a client to connect...")
	conn, err := (*listener).Accept()
	if err != nil {
		return conn, err
	}
	fmt.Printf("Accepted conn from %s\n", conn.RemoteAddr())
	return conn, err
}

func cutOff(client *Client) error {
	return (*client.conn).Close()
}

func sendMessage(client *Client, command ServerMessage) error {
	_, err := fmt.Fprintf(*client.conn, string(command))
	return err
}

func (client *Client) setName(name string) {
	client.Name = name
}
func (client *Client) getName() string {
	return client.Name
}

func (client *Client) setKeyIndex(keyID int) {
	client.KeyID = keyID
}
func (client *Client) getKeyIndex() int {
	return client.KeyID
}

func (client *Client) setHash(hash int) {
	client.Hash = hash
}

func (client *Client) getHash() int {
	return client.Hash
}
func getKeyPair(index int) KeyPair {
	switch index {
	case 0:
		return KeyPair{23019, 32037}
	case 1:
		return KeyPair{32037, 29295}
	case 2:
		return KeyPair{18789, 13603}
	case 3:
		return KeyPair{16443, 29533}
	case 4:
		return KeyPair{18189, 21952}
	default:
		return KeyPair{-1, -1}
	}
}

func countHashFromName(name string) int {
	var hash int
	for _, letter := range name {
		hash += int(letter)
	}
	hash *= 1000
	hash %= 65536
	return hash
}

func createConfirmationCode(clientHash int, key int) (confirmationCode int) {
	confirmationCode = clientHash
	confirmationCode += key
	confirmationCode %= 65536

	return confirmationCode
}

func codesMatch(code1 int, code2 int) bool {
	return code1 == code2
}

func handleClient(client *Client) {

	phase := USERNAME
	movePhase := DERIVE_POS
	buffer := ""
	for {
		// Read data from client
		data := make([]byte, 1024)
		n, err := (*client.conn).Read(data)
		if err != nil {
			//possible more exit point/non-standard situations?
			log.Println("Error reading data:", err)
			cutOff(client)
			break
		}

		// Append data to buffer
		buffer += string(data[:n])

		// Check if buffer contains a complete message
		for {
			index := strings.Index(buffer, TERMINATOR)
			if index == -1 {
				//check if we can early exit the program if error occurs
				switch phase {
				case USERNAME:
					if len(buffer) > MAX_NAME_LEN {
						sendMessage(client, SERVER_SYNTAX_ERROR)
						cutOff(client)
						return
					}
				case VALIDATION:
					codeAsNumber, _ := strconv.Atoi(buffer)
					if len(buffer) > 5 || codeAsNumber > 65536 {
						sendMessage(client, SERVER_SYNTAX_ERROR)
						cutOff(client)
					}
				}
				break // Incomplete message in buffer, wait for more data
			}

			// Extract complete message from buffer
			message := buffer[:index]
			buffer = buffer[index+2:]

			// Process complete message
			fmt.Printf("Received message: %s\n", message)

			switch phase {
			case MOVE:
				client.pos.x = 0
				client.pos.y = 0
				switch movePhase {
				case STRAIGHT:
					sendMessage(client, SERVER_MOVE)
				case RIGHT:
					sendMessage(client, SERVER_TURN_RIGHT)
				case LEFT:
					sendMessage(client, SERVER_TURN_LEFT)
				case DERIVE_POS:
					sendMessage(client, SERVER_MOVE)
					_, err := (*client.conn).Read(data)
					if err != nil {
						//possible more exit point/non-standard situations?
						log.Println("Error reading data:", err)
						cutOff(client)
						return
					}

				}

			case USERNAME:
				if len(message) > MAX_NAME_LEN {
					sendMessage(client, SERVER_SYNTAX_ERROR)
					cutOff(client)
					break
				}
				client.setName(message)
				sendMessage(client, SERVER_KEY_REQUEST)
				phase = KEY

			case KEY:
				keyID, err := strconv.Atoi(message)
				if err != nil {
					sendMessage(client, SERVER_SYNTAX_ERROR)
					cutOff(client)
					break
				}

				client.setKeyIndex(keyID)
				if keyID < MIN_KEY_INDEX || keyID > MAX_KEY_INDEX {
					sendMessage(client, SERVER_KEY_OUT_OF_RANGE_ERROR)
					cutOff(client)
					break
				}

				client.setHash(countHashFromName(client.getName()))
				serverConfirmationCode := createConfirmationCode(client.getHash(), getKeyPair(client.getKeyIndex()).ServerKey)

				stringHash := strconv.Itoa(serverConfirmationCode) + "\a\b"
				sendMessage(client, ServerMessage(stringHash))

				phase = VALIDATION

			case VALIDATION:
				clientConfirmationCode, err := strconv.Atoi(message)
				if err != nil || clientConfirmationCode > 65535 {
					sendMessage(client, SERVER_SYNTAX_ERROR)
					cutOff(client)
					break
				}
				validationCode := createConfirmationCode(client.getHash(), getKeyPair(client.getKeyIndex()).ClientKey)

				if !codesMatch(validationCode, clientConfirmationCode) {
					sendMessage(client, SERVER_LOGIN_FAILED)
					cutOff(client)
					break
				}
				sendMessage(client, SERVER_OK)
				phase = MOVE
			}
		}
	}
}

func main() {

	listener, err := createListener()
	if err != nil {
		fmt.Printf("failed to create listener: %s\n", err.Error())
		return
	}

	// cutOff only when the main function ends
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			return
		}
	}(listener)

	for {
		conn, err := waitForClientConnection(&listener)
		if err != nil {
			fmt.Printf("failed to accept client: %s\n", err.Error())
			continue
		}
		err = conn.SetDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			return
		}
		client := Client{conn: &conn}
		go handleClient(&client)
	}
}
