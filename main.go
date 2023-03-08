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

	SERVER_KEY_REQUEST     = "107 KEY REQUEST\a\b"
	UNABLE_TO_CLOSE_SOCKET = "unable to close connection\n"
)

type KeyPair struct {
	ServerKey string
	ClientKey string
}

func errorOccurred(err error, message string) bool {
	if err == nil {
		return false
	}
	fmt.Printf("%s: %s\n", err, message)
	return true
}

func closeSocket(connection *net.Conn) {
	err := (*connection).Close()
	errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
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

	//try to create socket
	connection, err := net.Dial(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if errorOccurred(err, "unable to create socket connection") {
		return
	}
	defer func(connection net.Conn) {
		closeSocket(&connection)
	}(connection)

	//setup connection timeout
	err = connection.SetDeadline(time.Now().Add(5 * time.Second))
	if errorOccurred(err, "unable to set timeout") {
		return
	}

	err = sendUsername(&connection, SAMPLE_USERNAME)
	if errorOccurred(err, "failed to send username") {
		closeSocket(&connection)
		return
	}

	////wait for key id request from server
	keyIDRequest, err := readKeyIDRequest(&connection)
	if errorOccurred(err, "failed to read key ID") {
		closeSocket(&connection)
		return
	}

	err = validateKeyIDRequest(keyIDRequest)
	if errorOccurred(err, "") {
		closeSocket(&connection)
		return
	}

	//send key id number
	err = sendKeyID("3", &connection)
	if err != nil {
		fmt.Printf("%s", err)
	}
}
