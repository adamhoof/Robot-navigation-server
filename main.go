package main

import (
	"fmt"
	"net"
)

const (
	SAMPLE_USERNAME = "robot1"
	SERVER_ADDRESS  = "localhost"
	SERVER_PORT     = "3999"

	PROTOCOL = "tcp"
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

func main() {
	//register key pairs
	availableKeyPairs := make([]KeyPair, 5)
	FillKeyPairs(availableKeyPairs)

	client, err := net.Dial(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer func(connection net.Conn) {
		err := connection.Close()
		if err != nil {

		}
	}(client)
}
