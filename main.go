package main

import (
	"fmt"
	"net"
)

const (
	SERVER_ADDRESS = "localhost"
	SERVER_PORT    = "3999"
	PROTOCOL       = "tcp"
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

func main() {
	//register key pairs
	availableKeyPairs := make([]KeyPair, 5)
	FillKeyPairs(availableKeyPairs)

	// Create a listener for incoming connections
	//prevent IPv6 incorrect host input with JoinHostPort()
	fmt.Printf("Starting server on %s:%s\n", SERVER_ADDRESS, SERVER_PORT)
	listener, err := net.Listen(PROTOCOL, net.JoinHostPort(SERVER_ADDRESS, SERVER_PORT))
	if err != nil {
		return
	}
	fmt.Println("Server started...")

	//close only when the main function ends
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	for {
		// Wait for a client to connect
		fmt.Println("Waiting for a client to connect...")
		client, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Accepted client from %s\n", client.RemoteAddr())
		//TODO: go handleClient
	}
}
