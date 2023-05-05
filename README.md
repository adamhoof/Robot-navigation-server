This project contains the server-side code for a multi-client game, where clients navigate through an environment to reach a certain target point. The server is written in Go and supports multiple concurrent clients, handling their communication and game logic.

The server consists of a main function that sets up a listener for incoming client connections. Upon a new connection, a new goroutine is spawned to handle client communication using the handleClient function.

The game logic is implemented in the handleClient and handleSingleMessage functions. The handleClient function reads incoming messages from the client, processes them, and generates responses. The incoming messages are handled according to the client's current phase (USERNAME, KEY, VALIDATION, MOVE, RECHARGING, and WIN).

The handleSingleMessage function processes individual messages based on the client's current phase and returns a response message and the next phase for the client.
Project Structure
