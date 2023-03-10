package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_ADDRESS         = "localhost"
	SERVER_PORT            = "3999"
	PROTOCOL               = "tcp"
	UNABLE_TO_CLOSE_SOCKET = "unable to close connection\n"

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
	connection         *net.Conn
	Name               string
	KeyID              int
	NameTerminatorPos  int
	KeyIDTerminatorPos int
}

func errorOccurred(err error, message string) bool {
	if err == nil {
		return false
	}
	fmt.Printf("%s: %s\n", err, message)
	return true
}

func closeSocket(client *Client) {
	err := (*client.connection).Close()
	errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
}

func FillKeyPairs() (keyPairs []KeyPair) {
	keyPairs = append(keyPairs, KeyPair{
		ServerKey: 23019,
		ClientKey: 32037,
	})

	keyPairs = append(keyPairs, KeyPair{
		ServerKey: 32037,
		ClientKey: 29295,
	})

	keyPairs = append(keyPairs, KeyPair{
		ServerKey: 18789,
		ClientKey: 13603,
	})

	keyPairs = append(keyPairs, KeyPair{
		ServerKey: 16443,
		ClientKey: 29533,
	})

	keyPairs = append(keyPairs, KeyPair{
		ServerKey: 18189,
		ClientKey: 21952,
	})
	return keyPairs
}

func writeToClient(client *Client, message string) error {
	_, err := fmt.Fprintf(*client.connection, message)
	return err
}

func readName(client *Client) (name string, err error) {
	buffer := make([]byte, 20)
	_, err = (*client.connection).Read(buffer)
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
	_, err = (*client.connection).Read(buffer)
	stringKeyID := string(buffer)
	client.KeyIDTerminatorPos = strings.Index(stringKeyID, "\a\b")

	keyID, err = strconv.Atoi(stringKeyID[:client.KeyIDTerminatorPos])
	//returns int and possibly error
	return
}

func countHash(client *Client, keyPairs []KeyPair) int {
	var hash int
	for i := 0; i < client.NameTerminatorPos; i++ {
		hash += int(client.Name[i])
	}
	hash *= 1000
	hash %= 65536
	hash += keyPairs[(client.KeyID)].ServerKey
	hash %= 65536
	return hash
}

func main() {
	//register key pairs
	/*availableKeyPairs := FillKeyPairs()*/

	// Create a listener for incoming connections
	//prevent IPv6 incorrect host input with JoinHostPort()
	fmt.Printf("Starting server on %s:%s\n", SERVER_ADDRESS, SERVER_PORT)
	listener, err := net.Listen(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if err != nil {
		fmt.Printf("failed to start server: %s\n", err)
		return
	}
	fmt.Println("Server started...")

	//close only when the main function ends
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Printf("unable to close listener: %s\n", err)
			return
		}
	}(listener)

	for {
		//TODO WRAPPER

		// Wait for a client to connect
		log.Println("Waiting for a client to connect...")
		conn, err := listener.Accept()
		if errorOccurred(err, "failed to accept socket communication") {
			continue
		}
		fmt.Printf("Accepted conn from %s\n", conn.RemoteAddr())
		/*client := Client{connection: &conn}*/

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return
		}

		messageContent := make([]byte, 10000)
		_, err = conn.Read(messageContent)
		if err != nil {
			return
		}

		cleanedUpMessage := strings.TrimRight(string(messageContent), "\x00")
		splitMessagesArray := strings.Split(cleanedUpMessage, "\a\b")

		lastIndex := len(splitMessagesArray) - 1
		if splitMessagesArray[lastIndex] == "" {
			splitMessagesArray = splitMessagesArray[:lastIndex]
		}
		for _, message := range splitMessagesArray {
			fmt.Printf("Received message: %s\n", message)
		}
		//wait for client to send name
		/*client.Name, err = readName(scanner)
		fmt.Println(client.Name)
		if errorOccurred(err, "failed to read name") {
			closeSocket(&client)
			continue
		}

		//position return
		client.NameTerminatorPos, err = checkValidityOfName(client.Name)
		if errorOccurred(err, "") {
			_ = writeToClient(&client, err.Error())
			closeSocket(&client)
			continue
		}

		err = requestKeyID(&client)
		if errorOccurred(err, "unable to request key id") {
			closeSocket(&client)
			continue
		}

		//wait for client to send key id number
		client.KeyID, err = readKeyID(&client)
		if errorOccurred(err, "unable to read key id") {
			closeSocket(&client)
			continue
		}
		fmt.Printf("client id: %d\n", client.KeyID)

		hash := countHash(&client, availableKeyPairs)
		fmt.Printf("hash: %d\n", hash)*/
	}
}
