package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	SERVER_ADDRESS = "localhost"
	SERVER_PORT    = "3999"
	PROTOCOL       = "tcp"

	SERVER_KEY_REQUEST = "107 KEY REQUEST\a\b"

	UNABLE_TO_CLOSE_SOCKET = "unable to close connection\n"
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

func readName(client *Client) (name string, err error) {
	buffer := make([]byte, 24)
	_, err = (*client.connection).Read(buffer)
	name = string(buffer)
	return name, err
}
func checkValidityOfName(name string) (terminatorPosition int, err error) {
	terminatorPosition = strings.Index(name, "\a\b")
	if terminatorPosition == -1 {
		return terminatorPosition, fmt.Errorf("terminator not found")
	}
	if terminatorPosition > 20 {
		return terminatorPosition, fmt.Errorf("invalid length of name")
	}
	return terminatorPosition, err
}

func requestKeyID(client *Client) error {
	_, err := fmt.Fprintf(*client.connection, SERVER_KEY_REQUEST)
	return err
}

func readKeyID(client *Client) (keyID int, err error) {
	buffer := make([]byte, 10)
	_, err = (*client.connection).Read(buffer)
	stringKeyID := string(buffer)
	client.KeyIDTerminatorPos = strings.Index(stringKeyID, "\a\b")

	//returns int and possibly error
	return strconv.Atoi(stringKeyID[:client.KeyIDTerminatorPos])
}

func countHash(client *Client, keyPairs []KeyPair) int {
	var hash int
	for i := 0; i < client.NameTerminatorPos; i++ {
		hash += int(client.Name[i])
	}
	hash *= 1000
	hash %= 65536
	hash += keyPairs[(client.KeyID)-1].ServerKey
	hash %= 65536
	return hash
}

func main() {
	//register key pairs
	availableKeyPairs := FillKeyPairs()

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
		// Wait for a client to connect
		fmt.Println("Waiting for a connection to connect...")
		connection, err := listener.Accept()
		if errorOccurred(err, "failed to accept socket communication") {
			continue
		}
		fmt.Printf("Accepted connection from %s\n", connection.RemoteAddr())
		client := Client{connection: &connection}

		//wait for client to send name
		client.Name, err = readName(&client)
		fmt.Println(client.Name)
		if errorOccurred(err, "failed to read name") {
			closeSocket(&client)
			continue
		}

		//position return
		client.NameTerminatorPos, err = checkValidityOfName(client.Name)
		if errorOccurred(err, "") {
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
		fmt.Printf("hash: %d\n", hash)
	}
}
