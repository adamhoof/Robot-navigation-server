package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	SERVER_ADDRESS = "localhost"
	SERVER_PORT    = "3999"
	PROTOCOL       = "tcp"

	TERMINATOR = "\a\b"

	MAX_NAME_LEN           = 18
	MIN_KEY_INDEX          = 0
	MAX_KEY_INDEX          = 4
	UNABLE_TO_CLOSE_SOCKET = "unable to cutOff conn\n"

	SERVER_KEY_REQUEST = "107 KEY REQUEST\a\b"

	SERVER_MOVE = "102 MOVE\a\b"

	SERVER_TURN_LEFT  = "103 TURN LEFT\a\b"
	SERVER_TURN_RIGHT = "104 TURN RIGHT\a\b"
	SERVER_PICK_UP    = "105 GET MESSAGE\a\b"
	SERVER_LOGOUT     = "106 TURN LEFT\a\b"

	SERVER_OK                     = "200 OK\a\b"
	SERVER_LOGIN_FAILED           = "300 LOGIN FAILED\a\b"
	SERVER_SYNTAX_ERROR           = "301 SYNTAX ERROR\a\b"
	SERVER_LOGIC_ERROR            = "302 LOGIC ERROR\a\b"
	SERVER_KEY_OUT_OF_RANGE_ERROR = "303 KEY OUT OF RANGE\a\b"
)

type KeyPair struct {
	ServerKey int
	ClientKey int
}

type Client struct {
	conn               *net.Conn
	Name               string
	KeyID              int
	NameTerminatorPos  int
	KeyIDTerminatorPos int
}

func (client *Client) getName() string {
	return client.Name
}

func (client *Client) setName(name string) {
	client.Name = name
}

func (client *Client) getKeyIndex() int {
	return client.KeyID
}

func (client *Client) setKeyIndex(keyID int) {
	client.KeyID = keyID
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

func sendMessage(client *Client, message string) error {
	_, err := fmt.Fprintf(*client.conn, message)
	return err
}

func cutOff(client *Client) error {
	return (*client.conn).Close()
}

func countHash(client *Client, keyPair KeyPair) int {
	var hash int
	for _, letter := range client.getName() {
		hash += int(letter)
	}
	hash *= 1000
	hash %= 65536
	hash += keyPair.ServerKey
	hash %= 65536
	return hash
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

func handleClient(client *Client) {

	phase := "username"

	buffer := ""
	for {
		// Read data from client
		data := make([]byte, 1024)
		n, err := (*client.conn).Read(data)
		if err != nil {
			sendMessage(client, SERVER_KEY_REQUEST)
			log.Println("Error reading data:", err)
			break
		}

		// Append data to buffer
		buffer += string(data[:n])

		// Check if buffer contains a complete message
		for {
			index := strings.Index(buffer, TERMINATOR)
			if index == -1 {
				break // Incomplete message in buffer, wait for more data
			}

			// Extract complete message from buffer
			message := buffer[:index]
			buffer = buffer[index+2:]

			// Process complete message
			fmt.Printf("Received message: %s\n", message)

			switch phase {

			case "username":
				if len(message) > MAX_NAME_LEN {
					sendMessage(client, SERVER_SYNTAX_ERROR)
					cutOff(client)
					break
				}
				client.setName(message)
				sendMessage(client, SERVER_KEY_REQUEST)
				phase = "key"

			case "key":
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

				hash := countHash(client, getKeyPair(client.getKeyIndex()))
				stringHash := strconv.Itoa(hash) + "\a\b"

				sendMessage(client, stringHash)
				phase = "confirmation"

			case "confirmation":
				sendMessage(client, SERVER_OK)
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
		client := Client{conn: &conn}
		go handleClient(&client)
	}
}
