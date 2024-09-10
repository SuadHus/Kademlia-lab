package kademlia

import (
	"fmt"
	"net"
	"strings"
)

// Network struct definition
type Network struct {
	LocalID   *KademliaID // Exported by making the first letter uppercase
	LocalAddr string      // Exported by making the first letter uppercase
}

func Listen(ip string, port int) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting the listener:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go parseConnection(conn)

		// shall handle ping etc...
		go handleConnection(conn)
	}
}

func parseConnection(conn net.Conn) {
	defer conn.Close()

	// Create a buffer to store the incoming data
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// convert buffer to string and trim whithespaces
	message := string(buffer[:n])
	message = strings.TrimSpace(message)

	switch {
	case strings.HasPrefix(message, "PING"):
		fmt.Println("Received PING message:", message)
		handlePingConnection(conn) // Delegate to handlePingConnection for PING

	case strings.HasPrefix(message, "HELLO"):
		fmt.Println("Received HELLO message:", message)
		// Add logic for handling HELLO messages here

	case strings.HasPrefix(message, "BYE"):
		fmt.Println("Received BYE message:", message)
		// Add logic for handling BYE messages here

	default:
		fmt.Println("Received unknown message:", message)
		// Handle unknown message types here
	}
}

// handleConnection handles incoming messages
func handleConnection(conn net.Conn) {
	fmt.Println("Handling connection from:", conn.RemoteAddr())
	defer conn.Close()
	// Handle incoming messages here TODO LATER
}

func handlePingConnection(conn net.Conn) {
	localAddr := conn.LocalAddr().String()

	// PONG answer with the recipient ip of the original PING
	pongMessage := "PONG from:" + localAddr
	_, err := conn.Write([]byte(pongMessage))
	if err != nil {
		fmt.Println("Error sending PONG response:", err)
		return
	}

	fmt.Println("Sent PONG message with server IP:", localAddr)
}

// SendPingMessage sends a ping message to the given contact
func (network *Network) SendPingMessage(contact *Contact) {
	conn, err := net.Dial("tcp", contact.Address)
	if err != nil {
		fmt.Println("Error connecting to contact:", err)
		return
	}
	defer conn.Close()

	message := fmt.Sprintf("PING from %s", network.LocalID.String())
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending ping message:", err)
	}
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
