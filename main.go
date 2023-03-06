package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	SAMPLE_USERNAME = "Mnau!\a\b"
	SERVER_ADDRESS  = "localhost"
	SERVER_PORT     = "3999"

	PROTOCOL = "tcp"

	SERVER_KEY_REQUEST = "107 KEY REQUEST\a\b"
)

type KeyPair struct {
	ServerKey string
	ClientKey string
}

// YEET EM' IN
func FillKeyPairs(KeyPairs []KeyPair) {
	KeyPairs = append(KeyPairs, KeyPair{
		ServerKey: "23019",
		ClientKey: "32037",
	})

	KeyPairs = append(KeyPairs, KeyPair{
		ServerKey: "32037",
		ClientKey: "29295",
	})

	KeyPairs = append(KeyPairs, KeyPair{
		ServerKey: "18789",
		ClientKey: "13603",
	})

	KeyPairs = append(KeyPairs, KeyPair{
		ServerKey: "16443",
		ClientKey: "29533",
	})

	KeyPairs = append(KeyPairs, KeyPair{
		ServerKey: "18189",
		ClientKey: "21952",
	})
}

func sendUsername(conn *net.Conn, username string) error {
	_, err := fmt.Fprintf(*conn, username+"\a\b")
	return err
}

func readKeyIDRequest(connection *net.Conn) (keyID string, err error) {
	buffer := make([]byte, 1024)
	_, err = (*connection).Read(buffer)
	keyID = string(buffer)
	return keyID, err
}

func validateKeyIDRequest(request string) (err error) {
	index := strings.Index(request, "\a\b")
	if index == -1 {
		return fmt.Errorf("invalid format")
	}
	diff := strings.Compare(request[:index+2], SERVER_KEY_REQUEST)
	if diff != 0 {
		return fmt.Errorf("invalid key id request")
	}
	return err
}

func sendKeyID(keyID string, connection *net.Conn) error {
	id := []byte(keyID + "\a\b")
	_, err := (*connection).Write(id)
	return err
}

func main() {
	//register key pairs
	availableKeyPairs := make([]KeyPair, 5)
	FillKeyPairs(availableKeyPairs)

	connection, err := net.Dial(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer func(connection net.Conn) {
		err := connection.Close()
		if err != nil {
		}
	}(connection)

	err = connection.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		fmt.Printf("unable to set timeout: %s", err)
	}

	err = sendUsername(&connection, SAMPLE_USERNAME)
	if err != nil {
		fmt.Printf("failed to send username: %s", err)
	}

	keyIDRequest, err := readKeyIDRequest(&connection)
	if err != nil {
		fmt.Printf("failed to read key ID: %s", err)
	}
	err = validateKeyIDRequest(keyIDRequest)
	if err != nil {
		fmt.Printf("%s", err)
	}

	err = sendKeyID("3", &connection)
	if err != nil {
		fmt.Printf("%s", err)
	}
}
