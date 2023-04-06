package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const (
	SERVER_ADDRESS = "localhost"
	SERVER_PORT    = "3999"
	PROTOCOL       = "tcp"
)

const (
	TERMINATOR   = "\a\b"
	MAX_NAME_LEN = 18
	MIN_KEY_ID   = 0
	MAX_KEY_ID   = 4
)

type ServerMessage string

const (
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
	WIN              Phase = 4
)

type Direction int

const (
	UNKNOWN Direction = -1

	UP   Direction = 0
	DOWN Direction = 1
	R    Direction = 2
	L    Direction = 3

	DIR_STRAIGHT Direction = 10
)

type KeyPair struct {
	ServerKey int
	ClientKey int
}

type Client struct {
	conn          *net.Conn
	Name          string
	KeyID         int
	Hash          int
	phase         Phase
	lastMovePhase MovePhase
	movePhase     MovePhase
	pos           Position
	lastPos       Position
	dir           Direction
	targetDir     Direction
	facing        Direction
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

func (client *Client) getXPos() int {
	return client.pos.x
}

func (client *Client) getYPos() int {
	return client.pos.y
}

func (client *Client) getLastXPos() int {
	return client.lastPos.x
}

func (client *Client) getLastYPos() int {
	return client.lastPos.y
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

func currentDirection(position Position, lastPosition Position) Direction {
	if position.x == lastPosition.x {
		if position.y < lastPosition.y {
			return DOWN
		}
		return UP

	} else if position.y == lastPosition.y {
		if position.x < lastPosition.x {
			return L
		}
		return R
	}
	//should never happen
	return UNKNOWN
}

func upperRightQuadrant(position Position) bool {
	return position.y >= 0 && position.x > 0
}

func downRightQuadrant(position Position) bool {
	return position.y <= 0 && position.x > 0
}

func upperLeftQuadrant(position Position) bool {
	return position.y >= 0 && position.x < 0
}

func downLeftQuadrant(position Position) bool {
	return position.y <= 0 && position.x < 0
}
func upperMidQuadrant(position Position) bool {
	return position.y > 0 && position.x == 0
}
func downMidQuadrant(position Position) bool {
	return position.y < 0 && position.x == 0
}
func calibrateDirection(direction Direction, position Position) (movementDir Direction, facing Direction) {

	switch direction {
	case UP:
		if upperRightQuadrant(position) {
			return L, L
		} else if upperLeftQuadrant(position) {
			return R, R
		}
		// upperRight => L
		// downRight => S
		// upperLeft => R
		// downLeft => S
		//upperMid =>  should never happen
		//downMid => S

		//downMidQuadrant, utopian
	case DOWN:
		if downRightQuadrant(position) {
			return R, L
		} else if downLeftQuadrant(position) {
			return L, R
		}
		// upperRight => S
		// downRight => R
		// upperLeft => S
		// downLeft => L
		//upperMid =>  S
		//downMid => should never happen

		//upperLeftQuadrant or upperMidQuadrant or upperRightQuadrant, utopian
	case R:
		if upperRightQuadrant(position) || upperMidQuadrant(position) {
			return R, DOWN
		} else if downRightQuadrant(position) || downMidQuadrant(position) {
			return L, UP
		}
		//upperLeftQuadrant or downLeftQuadrant, utopian
	case L:
		if upperLeftQuadrant(position) || upperMidQuadrant(position) {
			return L, DOWN
		} else if downLeftQuadrant(position) || downMidQuadrant(position) {
			return R, UP
		}
		//upperRightQuadrant or downRightQuadrant, utopian
	}
	return DIR_STRAIGHT, direction
}

func positionChanged(pos Position, lastPos Position) bool {
	return !(pos.x == lastPos.x && pos.y == lastPos.y)
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
		var err error
		client.lastPos = client.pos
		client.pos, err = extractPosition(singleMessage)
		if err != nil {
			return SERVER_SYNTAX_ERROR, CLOSE_CONNECTION
		}
		if isCenterPos(&client.pos) {
			return SERVER_PICK_UP, WIN
		}

		switch client.movePhase {
		case LOCATE:
			client.movePhase = CALIBRATE
			return SERVER_MOVE, MOVE

		case CALIBRATE:
			if !positionChanged(client.pos, client.lastPos) {
				client.movePhase = LOCATE

				switch client.facing {
				case UP:
					if upperRightQuadrant(client.pos) || downRightQuadrant(client.pos) {
						return SERVER_TURN_LEFT, MOVE
					} else if upperLeftQuadrant(client.pos) || downLeftQuadrant(client.pos) {
						return SERVER_TURN_RIGHT, MOVE
					}
					//if downMidQuadrant or UNKNOWN dir, just go left
					//upperMidQuadrant should not happen
				case DOWN:
					if upperRightQuadrant(client.pos) || downRightQuadrant(client.pos) {
						return SERVER_TURN_RIGHT, MOVE
					} else if upperLeftQuadrant(client.pos) || downLeftQuadrant(client.pos) {
						return SERVER_TURN_LEFT, MOVE
					}
				case R:
					if upperLeftQuadrant(client.pos) {
						return SERVER_TURN_RIGHT, MOVE
					} else if downLeftQuadrant(client.pos) {
						return SERVER_TURN_LEFT, MOVE
					}
				case L:
					if upperRightQuadrant(client.pos) || upperLeftQuadrant(client.pos) {
						return SERVER_TURN_LEFT, MOVE
					} else if downRightQuadrant(client.pos) || downLeftQuadrant(client.pos) {
						return SERVER_TURN_RIGHT, MOVE
					}
				}
				return SERVER_TURN_LEFT, MOVE
			}

			client.dir, client.facing = calibrateDirection(currentDirection(client.pos, client.lastPos), client.pos)

			switch client.dir {
			case DIR_STRAIGHT:
				client.movePhase = CALIBRATE
				return SERVER_MOVE, MOVE
			case R:
				client.movePhase = RIGHT
				return SERVER_TURN_RIGHT, MOVE
			case L:
				client.movePhase = LEFT
				return SERVER_TURN_LEFT, MOVE
			}
		case RIGHT:
			client.movePhase = CALIBRATE
			return SERVER_MOVE, MOVE
		case LEFT:
			client.movePhase = CALIBRATE
			return SERVER_MOVE, MOVE

		}
	case WIN:
		if len(singleMessage) > 98 {
			return SERVER_SYNTAX_ERROR, CLOSE_CONNECTION
		}
		return SERVER_LOGOUT, CLOSE_CONNECTION
	}
	return
}

type MovePhase int

const (
	LOCATE    MovePhase = 0
	CALIBRATE MovePhase = 1
	STRAIGHT  MovePhase = 2
	RIGHT     MovePhase = 3
	LEFT      MovePhase = 4

	UNDEF MovePhase = 30
)

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
		/*err = (*client.conn).SetDeadline(time.Now().Add(1 * time.Second))*/

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
			case MOVE:
				terminatorBeginningIndex := strings.Index(message, "\a")
				if terminatorBeginningIndex == -1 {
					continue
				}
				nonTerminatedMessage := message[:terminatorBeginningIndex]
				_, err = extractPosition(nonTerminatedMessage)
				if err != nil {
					sendMessage(client, SERVER_SYNTAX_ERROR)
					cutOff(client)
					return
				}
			case WIN:
				if len(message) > 99 {
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

				var response ServerMessage
				response, client.phase = handleSingleMessage(singleMessage, client)

				if client.phase == CLOSE_CONNECTION {
					sendMessage(client, response)
					cutOff(client)
					return
				}
				sendMessage(client, response)
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
		/*err = conn.SetDeadline(time.Now().Add(1 * time.Second))*/
		if err != nil {
			return
		}
		client := Client{conn: &conn, phase: USERNAME, movePhase: LOCATE, dir: UNKNOWN}
		go handleClient(&client)
	}
}
