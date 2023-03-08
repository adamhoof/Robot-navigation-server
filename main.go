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

	connection, err := net.Dial(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if errorOccurred(err, "unable to create socket connection") {
		return
	}
	defer func(connection net.Conn) {
		closeSocket(&connection)
	}(connection)

	err = connection.SetDeadline(time.Now().Add(5 * time.Second))
	if errorOccurred(err, "unable to set timeout") {
		return
	}

	//TODO if error close connection function wrapper
	err = sendUsername(&connection, SAMPLE_USERNAME)
	if errorOccurred(err, "failed to send username") {
		closeSocket(&connection)
		return
	}

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

	err = sendKeyID("3", &connection)
	if err != nil {
		fmt.Printf("%s", err)
	}
}

/*package main

import (
	"fmt"
	"net"
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
	ServerKey string
	ClientKey string
}

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

func readName(connection *net.Conn) (name string, err error) {
	buffer := make([]byte, 1024)
	_, err = (*connection).Read(buffer)
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

func requestKeyID(connection *net.Conn) error {
	_, err := fmt.Fprintf(*connection, SERVER_KEY_REQUEST)
	return err
}

func readKeyID(connection *net.Conn) (keyID string, err error) {
	buffer := make([]byte, 10)
	_, err = (*connection).Read(buffer)
	keyID = string(buffer)
	return keyID, err
}

func errorOccurred(err error, message string) bool {
	if err == nil {
		return true
	}
	fmt.Printf("%s: %s\n", err, message)
	return false
}
func main() {
	//register key pairs
	availableKeyPairs := make([]KeyPair, 5)
	FillKeyPairs(availableKeyPairs)

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
		// Wait for a connection to connect
		fmt.Println("Waiting for a connection to connect...")
		connection, err := listener.Accept()
		if errorOccurred(err, "failed to accept socket communication") {
			continue
		}
		fmt.Printf("Accepted connection from %s\n", connection.RemoteAddr())

		name, err := readName(&connection)
		fmt.Println(name)
		if errorOccurred(err, "failed to read name") {
			err = connection.Close()
			errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
			continue
		}

		//position return
		_, err = checkValidityOfName(name)
		if errorOccurred(err, "") {
			err = connection.Close()
			errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
			continue
		}

		err = requestKeyID(&connection)
		if errorOccurred(err, "unable to request key id") {
			err = connection.Close()
			errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
			continue
		}

		keyID, err := readKeyID(&connection)
		if errorOccurred(err, "unable to read key id") {
			err = connection.Close()
			errorOccurred(err, UNABLE_TO_CLOSE_SOCKET)
			continue
		}
		fmt.Println(keyID)
	}
}
*/
