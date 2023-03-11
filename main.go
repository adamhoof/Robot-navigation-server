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

func receivedMoreMessages(arrayOfMessages []string) bool {
	return len(arrayOfMessages) > 1
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

func handleClient(conn *net.Conn) {
	phase := "auth"
	err := (*conn).SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return
	}

	buffer := make([]byte, 10000)

	_, err = (*conn).Read(buffer)
	if err != nil {
		fmt.Println("failed to read data from client")
		return
	}

	messages := splitIntoMessages(buffer)

	//state machine
	if receivedMoreMessages(messages) {
		switch phase {
		case "auth":
		case "nav":
		case "recharging":

		}
		//continue continue doing more phases at once
	} else {
		//do one phase, then the other
	}
	for _, m := range messages {
		fmt.Println(m)
	}
}

func main() {
	//register key pairs
	availableKeyPairs := FillKeyPairs()
	fmt.Println(len(availableKeyPairs))

	listener, err := createListener()
	if err != nil {
		fmt.Printf("failed to create listener: %s\n", err.Error())
		return
	}

	//close only when the main function ends
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
		go handleClient(&conn)

		//go handleClient(&conn)
		/*client := Client{connection: &conn}*/
		/*err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return
		}

		buffer := make([]byte, 10000)

		_, err = conn.Read(buffer)
		if err != nil {
			fmt.Println("failed to read data from client")
			return
		}

		messages := splitIntoMessages(buffer)

		if receivedMoreMessages(messages) {
			//continue continue doing more phases at once
		} else {
			//do one phase, then the other
		}*/
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
