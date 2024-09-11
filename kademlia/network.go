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

// Listen starts a listener on the specified IP and port
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
		// Parse and handle the connection asynchronously
		go parseConnection(conn)
	}
}

func parseConnection(conn net.Conn) {
	defer conn.Close()

	// Create a buffer to store the incoming data
	buffer := make([]byte, 1024) // Adjust the buffer size as needed
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Convert the buffer to a string and trim any trailing whitespace
	message := string(buffer[:n])
	message = strings.TrimSpace(message)

	// Use a switch case to handle different message types
	switch {
	case strings.HasPrefix(message, "PING"):
		go handlePingMessage(conn, message)

	case strings.HasPrefix(message, "PONG"):
		fmt.Println("Received PONG message:", message)
		// Optionally handle the PONG message further

	case strings.HasPrefix(message, "BYE"):
		fmt.Println("Received BYE message:", message)
		// Add logic for handling BYE messages here

	default:
		fmt.Println("Received unknown message:", message)
		// Handle unknown message types here
	}
}

// handlePingMessage will process the PING message and respond with a PONG on a new connection
func handlePingMessage(conn net.Conn, message string) {
	fmt.Println("Received PING message from", conn.RemoteAddr())

	// Extract the IP and port from the remote connection
	remoteAddr := conn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		fmt.Println("Error extracting IP address from connection:", err)
		return
	}

	// Create a new connection to the sender to send a PONG message
	newConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, 8080)) // Assume PONG goes to port 9999 or similar
	if err != nil {
		fmt.Println("Error connecting back to sender:", err)
		return
	}
	defer newConn.Close()

	// Prepare and send the PONG message
	pongMessage := fmt.Sprintf("PONG from %s", newConn.LocalAddr().String())
	_, err = newConn.Write([]byte(pongMessage))
	if err != nil {
		fmt.Println("Error sending PONG message:", err)
		return
	}

	fmt.Println("Sent PONG message to", newConn.RemoteAddr().String())
}

// SendPingMessage sends a PING message to a given contact
func (network *Network) SendPingMessage(contact *Contact) {
	// Establish a connection to the contact
	conn, err := net.Dial("tcp", contact.Address)
	if err != nil {
		fmt.Println("Error connecting to contact:", err)
		return
	}
	defer conn.Close() // Keep connection open until after response is handled

	// Send the PING message
	message := fmt.Sprintf("PING from %s", network.LocalID.String())
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending ping message:", err)
		return
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
