/*
	TelChatNet -- A chat server based on Golang and Telnet.
	Run this like
		> go run ./main.go
	That will run a TCP chat server at localhost:9955.
	You can connect to that chat server like
		> telnet localhost 9955
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	WelcomeMsg = "" +
		"+-----------------------------------------------+\n" +
		"| Welcome to use TelChatNet!                    |\n" +
		"| If you have problems with this tools,         |\n" +
		"|please submit s issues here: AGDDoS/TelChatNet |\n" +
		"+-----------------------------------------------+"
)

func main() {
	// Number of people whom ever connected
	//
	clientCount := 0

	// All people who are connected; a map wherein the keys are net.Conn objects and the values are client "ids", an integer.
	//
	allClients := make(map[net.Conn]int)

	// Channel into which the TCP server will push new connections.
	//
	newConnections := make(chan net.Conn)

	// Channel into which we'll push dead connections for removal from allClients.
	//
	deadConnections := make(chan net.Conn)

	// Channel into which we'll push messages from connected clients so that we can broadcast them to every connection in allClients.
	//
	messages := make(chan string)

	// Start the TCP server
	//
	server, err := net.Listen("tcp", ":9955")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(WelcomeMsg)
	fmt.Println("[INIT]Server init Ok!")
	// Tell the server to accept connections forever and push new connections into the newConnections channel.
	//
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			newConnections <- conn
		}
	}()

	// Loop
	//
	for {

		// Handle 1) New connections; 2) Dead connections; 3) Broadcast messages.
		//
		select {

		// Accept new clients
		//
		case conn := <-newConnections:

			log.Printf("Accepted new client, #%d", clientCount)

			// Add this connection to the `allClients` map
			//
			allClients[conn] = clientCount
			clientCount += 1

			// Constantly read incoming messages from this client in a goroutine and push those onto the messages channel for broadcast to others.
			//
			go func(conn net.Conn, clientId int) {
				reader := bufio.NewReader(conn)
				for {
					incoming, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					messages <- fmt.Sprintf("Client %d > %s", clientId, incoming)
				}

				// When we encouter `err` reading, send this  connection to `deadConnections` for removal.
				//
				deadConnections <- conn

			}(conn, allClients[conn])

		// Accept messages from connected clients
		//
		case message := <-messages:

			// Loop over all connected clients
			//
			for conn := range allClients {

				// Send them a message in a go-routine so that the network operation doesn't block
				//
				go func(conn net.Conn, message string) {
					_, err := conn.Write([]byte(message))

					// If there was an error communicating with them, the connection is dead.
					if err != nil {
						deadConnections <- conn
					}
				}(conn, message)
			}
			log.Printf("New message: %s", message)
			log.Printf("Broadcast to %d clients", len(allClients))

		// Remove dead clients
		//
		case conn := <-deadConnections:
			log.Printf("Client %d disconnected", allClients[conn])
			delete(allClients, conn)
		}
	}
}
