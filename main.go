package main

import (
	"errors"
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
	UNABLE_TO_CLOSE_SOCKET = "unable to close conn\n"

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

func (client *Client) getKeyID() int {
	return client.KeyID
}

func (client *Client) setKeyID(keyID int) {
	client.KeyID = keyID
}

func (client *Client) close() error {
	return (*client.conn).Close()
}

func errorOccurred(err error, message string) bool {
	if err == nil {
		return false
	}
	fmt.Printf("%s: %s\n", err, message)
	return true
}

func closeSocket(client *Client) {
	err := (*client.conn).Close()
	errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
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

func writeToClient(client *Client, message string) error {
	_, err := fmt.Fprintf(*client.conn, message)
	return err
}

func readName(client *Client) (name string, err error) {
	buffer := make([]byte, 20)
	_, err = (*client.conn).Read(buffer)
	if err != nil {
		return name, errors.New(SERVER_SYNTAX_ERROR)
	}
	name = string(buffer)
	return name, err
}
func checkValidityOfName(name string) (terminatorPosition int, err error) {
	terminatorPosition = strings.Index(name, "\a\b")
	if terminatorPosition == -1 || terminatorPosition > 20 {
		return terminatorPosition, errors.New(SERVER_SYNTAX_ERROR)
	}
	return terminatorPosition, err
}

func requestKeyID(client *Client) error {
	return writeToClient(client, SERVER_KEY_REQUEST)
}

func readKeyID(client *Client) (keyID int, err error) {
	buffer := make([]byte, 10)
	_, err = (*client.conn).Read(buffer)
	stringKeyID := string(buffer)
	client.KeyIDTerminatorPos = strings.Index(stringKeyID, "\a\b")

	keyID, err = strconv.Atoi(stringKeyID[:client.KeyIDTerminatorPos])
	//returns int and possibly error
	return
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

func receivedMoreMessages(arrayOfMessages []string) bool {
	return len(arrayOfMessages) > 1
}

func waitForClientMessage(client *Client) (buffer []byte, err error) {
	_, err = (*client.conn).Read(buffer)
	return buffer, err
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

func splitIntoMessages(message []byte) []string {
	cleanedUpMessage := strings.TrimRight(string(message), "\x00")
	splitMessagesArray := strings.Split(cleanedUpMessage, "\a\b")

	lastIndex := len(splitMessagesArray) - 1
	if splitMessagesArray[lastIndex] == "" {
		splitMessagesArray = splitMessagesArray[:lastIndex]
	}
	return splitMessagesArray
}

func handleClient(client *Client) {

	phase := "username"
	/*currentState := 0
	stateMap := make(map[int]string)
	stateMap[currentState] = "username"
	stateMap[currentState] = "key"
	stateMap[currentState] = "confirmation"*/

	fmt.Println("receiving communication")
	buffer := ""
	for {
		// Read data from client
		data := make([]byte, 1024)
		n, err := (*client.conn).Read(data)
		if err != nil {
			writeToClient(client, SERVER_KEY_REQUEST)
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

			/*phase := stateMap[currentState]*/
			switch phase {

			case "username":
				if len(message) > MAX_NAME_LEN {
					writeToClient(client, SERVER_SYNTAX_ERROR)
					client.close()
					break
				}
				client.setName(message)
				writeToClient(client, SERVER_KEY_REQUEST)
				/*currentState++*/
				phase = "key"

			case "key":
				keyID, err := strconv.Atoi(message)
				if err != nil {
					writeToClient(client, SERVER_SYNTAX_ERROR)
					client.close()
					break
				}

				client.setKeyID(keyID)
				if keyID < MIN_KEY_INDEX || keyID > MAX_KEY_INDEX {
					writeToClient(client, SERVER_KEY_OUT_OF_RANGE_ERROR)
					client.close()
					break
				}

				hash := countHash(client, getKeyPair(client.getKeyID()))
				stringHash := strconv.Itoa(hash) + "\a\b"
				writeToClient(client, stringHash)
				phase = "confirmation"
				/*currentState++*/
			case "confirmation":
				writeToClient(client, SERVER_OK)
			}
		}
	}
	cock := []byte{1, 2, 3}
	_, err := (*client.conn).Write(cock)
	if err != nil {
		return
	}
	fmt.Printf("cock")

}

func main() {

	listener, err := createListener()
	if err != nil {
		fmt.Printf("failed to create listener: %s\n", err.Error())
		return
	}

	// close only when the main function ends
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
