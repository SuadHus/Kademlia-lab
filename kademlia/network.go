package kademlia

import (
	"fmt"
	"net"
	"strings"
)

type Network struct {
	LocalID   *KademliaID
	LocalAddr string
	msgChan   chan string
}

// Listen starts a listener on the specified IP and port
func (network *Network) Listen(ip string, port int) {
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
		go network.parseConnection(conn)
	}
}

func (network *Network) parseConnection(conn net.Conn) {
	defer conn.Close()

	// Create a buffer to store the incoming data
	buffer := make([]byte, 1024)
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
		go network.handlePingMessage(conn, message)

	case strings.HasPrefix(message, "PONG"):
		fmt.Println("Received PONG message:", message)

	case strings.HasPrefix(message, "BYE"):
		fmt.Println("Received BYE message:", message)

	default:
		fmt.Println("Received unknown message:", message)
	}
}

// handlePingMessage processes the PING message and responds with a PONG
func (network *Network) handlePingMessage(conn net.Conn, message string) {
	fmt.Println("Received PING message from", message)

	remoteAddr := conn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		fmt.Println("Error extracting IP address from connection:", err)
		return
	}

	// Create a new connection to send a PONG message
	newConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, 8080))
	if err != nil {
		fmt.Println("Error connecting back to sender:", err)
		return
	}
	defer newConn.Close()

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
	network.msgChan <- "Sending PING message to " + contact.Address
	conn, err := net.Dial("tcp", contact.Address)
	if err != nil {
		fmt.Println("Error connecting to contact:", err)
		return
	}
	defer conn.Close()

	message := fmt.Sprintf("PING from %s %s", network.LocalID.String(), network.LocalAddr)
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending ping message:", err)
		return
	}
}
