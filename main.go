package main

import (
	"fmt"
	"net"
)

func ConnectClient(host string, port string, communicationType string) (connection net.Conn, err error) {
	if communicationType != "tcp" {
		return connection, fmt.Errorf("invalid communication type, valid: tcp")
	}
	connection, err = net.Dial(communicationType, fmt.Sprintf("%s:%s", host, port))
	return connection, err
}

func SendMessage(connection *net.Conn, message string) (err error) {
	_, err = (*connection).Write([]byte(message))
	if err != nil {
		return fmt.Errorf("unable to send message, %s", err.Error())
	}
	return err
}

func WaitForResponse(connection *net.Conn, buffer []byte) (numOfChars int, err error) {
	numOfChars, err = (*connection).Read(buffer)
	return numOfChars, err
}

func main() {
	connection, err := ConnectClient("localhost", "8080", "tcp")

	if err != nil {
		println("failed to connect client")
		return
	}
	defer connection.Close()

	err = SendMessage(&connection, "fucktard")
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := make([]byte, 1024)
	numOfChars, err := WaitForResponse(&connection, buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Received:", string(buf[:numOfChars]))
}
