package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_ADDRESS = "localhost"
	SERVER_PORT    = "3999"
	PROTOCOL       = "tcp"
)

const (
	TERMINATOR = "\a\b"

	SPACE = ' '

	MAX_NAME_LEN = 18
	MIN_KEY_ID   = 0
	MAX_KEY_ID   = 4
)

type ServerMessage string

const (
	UNABLE_TO_CLOSE_SOCKET = "unable to cutOff conn\n"

	SERVER_KEY_REQUEST ServerMessage = "107 KEY REQUEST" + TERMINATOR

	SERVER_MOVE ServerMessage = "102 MOVE" + TERMINATOR

	SERVER_TURN_LEFT  ServerMessage = "103 TURN LEFT" + TERMINATOR
	SERVER_TURN_RIGHT ServerMessage = "104 TURN RIGHT" + TERMINATOR
	SERVER_PICK_UP    ServerMessage = "105 GET MESSAGE" + TERMINATOR
	SERVER_LOGOUT     ServerMessage = "106 LOGOUT" + TERMINATOR

	SERVER_OK                     ServerMessage = "200 OK" + TERMINATOR
	SERVER_LOGIN_FAILED           ServerMessage = "300 LOGIN FAILED" + TERMINATOR
	SERVER_SYNTAX_ERROR           ServerMessage = "301 SYNTAX ERROR" + TERMINATOR
	SERVER_LOGIC_ERROR            ServerMessage = "302 LOGIC ERROR" + TERMINATOR
	SERVER_KEY_OUT_OF_RANGE_ERROR ServerMessage = "303 KEY OUT OF RANGE" + TERMINATOR
)

type Phase int

const (
	CLOSE_CONNECTION Phase = -1
	USERNAME         Phase = 0
	KEY              Phase = 1
	VALIDATION       Phase = 2
	MOVE             Phase = 3
	END              Phase = 4
)

type MovePhase int

const (
	DERIVE_POS MovePhase = 0
	STRAIGHT   MovePhase = 1
	RIGHT      MovePhase = 2
	LEFT       MovePhase = 3
)

type Direction int

const (
	UNKNOWN Direction = -1
	UP      Direction = 0
	DOWN    Direction = 1
	R       Direction = 2
	L       Direction = 3

	AVOID_OBSTACLE = 4
)

type KeyPair struct {
	ServerKey int
	ClientKey int
}

type Client struct {
	conn      *net.Conn
	Name      string
	KeyID     int
	Hash      int
	phase     Phase
	movePhase MovePhase
	pos       Position
	lastPos   Position
	dir       Direction
	numMoves  uint16
}

type Position struct {
	x int
	y int
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

func cutOff(client *Client) error {
	return (*client.conn).Close()
}

func sendMessage(client *Client, command ServerMessage) error {
	_, err := fmt.Fprintf(*client.conn, string(command))
	return err
}

func (client *Client) setName(name string) {
	client.Name = name
}
func (client *Client) getName() string {
	return client.Name
}

func (client *Client) setKeyIndex(keyID int) {
	client.KeyID = keyID
}
func (client *Client) getKeyIndex() int {
	return client.KeyID
}

func (client *Client) setHash(hash int) {
	client.Hash = hash
}

func (client *Client) getHash() int {
	return client.Hash
}
func getKeyPair(index int) KeyPair {
	switch index {
	case 0:
		return KeyPair{23019, 32037}
	case 1:
		return KeyPair{32037, 29295}
	case 2:
		return KeyPair{18789, 13603}
	case 3:
		return KeyPair{16443, 29533}
	case 4:
		return KeyPair{18189, 21952}
	default:
		return KeyPair{-1, -1}
	}
}

func countHashFromName(name string) int {
	var hash int
	for _, letter := range name {
		hash += int(letter)
	}
	hash *= 1000
	hash %= 65536
	return hash
}

func createConfirmationCode(clientHash int, key int) (confirmationCode int) {
	confirmationCode = clientHash
	confirmationCode += key
	confirmationCode %= 65536

	return confirmationCode
}

func codesMatch(code1 int, code2 int) bool {
	return code1 == code2
}
func extractPosition(message string) (pos Position, err error) {
	pattern := "^OK (-?[0-9]+) (-?[0-9]+)$"
	matched, _ := regexp.MatchString(pattern, message)
	if !matched {
		return pos, errors.New("wrong coordinates format")
	}

	_, err = fmt.Sscanf(message, "OK %d %d", &pos.x, &pos.y)
	return pos, err
}

func isCenterPos(position *Position) bool {
	return position.x == 0 && position.y == 0
}

func initialMoveSequenceDone(numSteps uint16) bool {
	return numSteps >= 2
}

func unknownDirection(direction Direction) bool {
	return direction == UNKNOWN
}

func xAxisCloser(pos *Position) bool {
	return math.Abs(float64(pos.x)) <= math.Abs(float64(pos.y))
}

/*func movingAwayFromXAxis(position *Position, lastPosition *Position) bool {

}*/

type MessageType int

const (
	INCOMPLETE_MESSAGE            MessageType = 1
	SINGLE_MESSAGE                MessageType = 2
	MULTI_MESSAGE                 MessageType = 3
	SINGLE_AND_INCOMPLETE_MESSAGE MessageType = 4
	MULTI_AND_INCOMPLETE_MESSAGE  MessageType = 5
	BRUH_MESSAGE                  MessageType = 6
)

func deriveMessageType(message string, terminator string) MessageType {
	messageCount := strings.Count(message, terminator)
	endsWithTerminator := strings.HasSuffix(message, terminator)

	if messageCount == 0 {
		fmt.Println("incomplete: ", message)
		return INCOMPLETE_MESSAGE

	} else if messageCount == 1 {
		if endsWithTerminator {
			fmt.Println("single: ", message)
			return SINGLE_MESSAGE

		}
		fmt.Println("single and incomplete: ", message)
		return SINGLE_AND_INCOMPLETE_MESSAGE

	} else if messageCount >= 2 {
		if endsWithTerminator {
			fmt.Println("multi: ", message)
			return MULTI_MESSAGE
		}
		fmt.Println("multi and incomplete: ", message)
		return MULTI_AND_INCOMPLETE_MESSAGE
	}
	fmt.Println("bruh message: ", message)
	return BRUH_MESSAGE
}

func deriveDir(position *Position, lastPosition *Position) Direction {
	//moved up or down
	if position.x == lastPosition.x {
		if position.y == lastPosition.y {
			return AVOID_OBSTACLE
		} else if position.y < lastPosition.y {
			return DOWN
		} else {
			return UP
		}
		//moved right or left
	} else if position.y == lastPosition.y {
		if position.x == lastPosition.x {
			return AVOID_OBSTACLE
		} else if position.x < lastPosition.x {
			return L
		} else {
			return R
		}
	}
	//should never happen
	return UNKNOWN
}

func handleSingleMessage(singleMessage string, client *Client) (response ServerMessage, nextPhase Phase) {
	switch client.phase {

	case USERNAME:
		if len(singleMessage) > MAX_NAME_LEN {
			return SERVER_SYNTAX_ERROR, CLOSE_CONNECTION
		}
		client.setName(singleMessage)
		return SERVER_KEY_REQUEST, KEY

	case KEY:
		keyID, err := strconv.Atoi(singleMessage)
		if err != nil {
			return SERVER_SYNTAX_ERROR, CLOSE_CONNECTION
		}

		client.setKeyIndex(keyID)
		if keyID < MIN_KEY_ID || keyID > MAX_KEY_ID {
			return SERVER_KEY_OUT_OF_RANGE_ERROR, CLOSE_CONNECTION
		}

		client.setHash(countHashFromName(client.getName()))
		serverConfirmationCode := createConfirmationCode(client.getHash(), getKeyPair(client.getKeyIndex()).ServerKey)

		stringHash := strconv.Itoa(serverConfirmationCode) + TERMINATOR
		return ServerMessage(stringHash), VALIDATION

	case VALIDATION:
		clientConfirmationCode, err := strconv.Atoi(singleMessage)
		if err != nil || clientConfirmationCode > 65535 {
			return SERVER_SYNTAX_ERROR, CLOSE_CONNECTION
		}
		confirmationCode := createConfirmationCode(client.getHash(), getKeyPair(client.getKeyIndex()).ClientKey)

		if !codesMatch(confirmationCode, clientConfirmationCode) {
			return SERVER_LOGIN_FAILED, CLOSE_CONNECTION
		}
		sendMessage(client, SERVER_OK)

		return SERVER_MOVE, MOVE

	case MOVE:
		client.numMoves++

		var err error
		client.lastPos = client.pos
		client.pos, err = extractPosition(singleMessage)
		if err != nil {
			return SERVER_SYNTAX_ERROR, CLOSE_CONNECTION
		}
		if isCenterPos(&client.pos) {
			return SERVER_PICK_UP, END
		}
		if !initialMoveSequenceDone(client.numMoves) {
			return SERVER_MOVE, MOVE
		}
		if unknownDirection(client.dir) {
			client.dir = deriveDir(&client.pos, &client.lastPos)
		}
		if xAxisCloser(&client.pos) {
			if client.dir == UP || client.dir == DOWN {

			}
			/*if movingAwayFromXAxis(&client.pos, &client.lastPos) {

			}*/
		}

	case END:
		return SERVER_LOGOUT, CLOSE_CONNECTION
	}
	return
}

func handleClient(client *Client) {
	var buffer = make([]byte, 1024)
	var message string

	for {
		numChars, err := (*client.conn).Read(buffer)
		if err != nil {
			log.Println("Error reading buffer:", err)
			cutOff(client)
			return
			//possible more exit point/non-standard situations?
		}
		err = (*client.conn).SetDeadline(time.Now().Add(1 * time.Second))

		message += string(buffer[:numChars])
		terminatorIndex := strings.Index(message, TERMINATOR)

		messageType := deriveMessageType(message, TERMINATOR)

		switch messageType {
		case INCOMPLETE_MESSAGE:
			//check early exit availability
			switch client.phase {
			case USERNAME:
				if len(message) > MAX_NAME_LEN {
					sendMessage(client, SERVER_SYNTAX_ERROR)

					cutOff(client)
					return
				}
			case VALIDATION:
				codeAsNumber, _ := strconv.Atoi(message)
				if len(message) > 5 || codeAsNumber > 65536 {
					sendMessage(client, SERVER_SYNTAX_ERROR)
					cutOff(client)
					return
				}
			default:
				continue
			}

		case SINGLE_MESSAGE:
			nonTerminatedMessage := message[:terminatorIndex]

			var response ServerMessage
			response, client.phase = handleSingleMessage(nonTerminatedMessage, client)

			if client.phase == CLOSE_CONNECTION {
				sendMessage(client, response)
				cutOff(client)
				return
			}
			sendMessage(client, response)
			message = ""

		case MULTI_MESSAGE:
			for {
				terminatorIndex = strings.Index(message, TERMINATOR)
				if terminatorIndex == -1 {
					break
				}
				singleMessage := message[:terminatorIndex]
				message = message[terminatorIndex+2:]

				switch client.phase {
				case USERNAME:
					if len(singleMessage) > MAX_NAME_LEN {
						sendMessage(client, SERVER_SYNTAX_ERROR)
						cutOff(client)
						return
					}
					client.setName(singleMessage)
					sendMessage(client, SERVER_KEY_REQUEST)

					client.phase = KEY
				case KEY:
					keyID, err := strconv.Atoi(singleMessage)
					if err != nil {
						sendMessage(client, SERVER_SYNTAX_ERROR)
						cutOff(client)
						return
					}

					client.setKeyIndex(keyID)
					if keyID < MIN_KEY_ID || keyID > MAX_KEY_ID {
						sendMessage(client, SERVER_KEY_OUT_OF_RANGE_ERROR)
						cutOff(client)
						return
					}

					client.setHash(countHashFromName(client.getName()))
					serverConfirmationCode := createConfirmationCode(client.getHash(), getKeyPair(client.getKeyIndex()).ServerKey)

					terminatedStringHash := strconv.Itoa(serverConfirmationCode) + TERMINATOR
					sendMessage(client, ServerMessage(terminatedStringHash))

					client.phase = VALIDATION
				case VALIDATION:
					clientConfirmationCode, err := strconv.Atoi(singleMessage)
					if err != nil || clientConfirmationCode > 65535 {
						sendMessage(client, SERVER_SYNTAX_ERROR)
						cutOff(client)
						return
					}
					confirmationCode := createConfirmationCode(client.getHash(), getKeyPair(client.getKeyIndex()).ClientKey)

					if !codesMatch(confirmationCode, clientConfirmationCode) {
						sendMessage(client, SERVER_LOGIN_FAILED)
						cutOff(client)
						return
					}
					sendMessage(client, SERVER_OK)
					sendMessage(client, SERVER_MOVE)

					client.phase = MOVE

				case MOVE:
					//even if the client wanted, he can send only one message containing his coordinates within multi message
					client.pos, err = extractPosition(message)
					if err != nil {
						sendMessage(client, SERVER_SYNTAX_ERROR)
						cutOff(client)
						return
					}
				}
			}

		case SINGLE_AND_INCOMPLETE_MESSAGE:
			break
			//check if, based on phase, the message makes sense copy from INCOMPLETE_MESSAGE state - ONLY CHECK THE INCOMPLETE MESSAGE, WHETHER IT MAKES SENSE
		case MULTI_AND_INCOMPLETE_MESSAGE:
			break
			//check if, based on phase, the messages makes sense, copy from INCOMPLETE_MESSAGE state - ONLY CHECK THE INCOMPLETE MESSAGE, WHETHER IT MAKES SENSE
		default:
			//should not happen
			return
		}
	}
}

func main() {

	listener, err := createListener()
	if err != nil {
		fmt.Printf("failed to create listener: %s\n", err.Error())
		return
	}

	// cutOff only when the main function ends
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
		err = conn.SetDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			return
		}
		client := Client{conn: &conn, phase: USERNAME, movePhase: DERIVE_POS, dir: UNKNOWN, numMoves: 0}
		go handleClient(&client)
	}
}
