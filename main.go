package main

import (
	"fmt"
	"io"
	"net"
)

func Listen(port string, communicationType string) (listener net.Listener, err error) {
	listener, err = net.Listen(communicationType, port)
	return listener, err
}

func main() {
	// Listen for incoming connections on port 8080
	listener, err := Listen(":8080", "tcp")
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()

	fmt.Println("Server started, waiting for connections...")

	// Loop forever, waiting for connections
	for {
		// Wait for a connection
		client, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handleConnection(client)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("New client connected")
	// Make a buffer to hold incoming data
	buf := make([]byte, 1024)

	// Loop forever, reading from the connection
	for {
		// Read from the connection
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading:", err.Error())
			return
		}

		// Print the incoming data
		fmt.Println("Received:", string(buf[:n]))

		// Send a response back to the client
		conn.Write([]byte("fucktard\n"))
		break
	}
}
