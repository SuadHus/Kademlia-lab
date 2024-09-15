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
	handler   MessageHandler
}

// MessageHandler interface for handling messages
type MessageHandler interface {
	HandleMessage(message string, conn net.Conn)
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
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Convert the buffer to a string and trim any trailing whitespace
	message := string(buffer[:n])
	message = strings.TrimSpace(message)

	// Pass the message to the handler
	if network.handler != nil {
		network.handler.HandleMessage(message, conn)
	} else {
		fmt.Println("No message handler set")
	}
}

// SendMessage sends a message to a given address and waits for a response
func (network *Network) SendMessage(address string, message string) (string, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(message))
	if err != nil {
		return "", err
	}

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}
	response := string(buffer[:n])
	response = strings.TrimSpace(response)
	return response, nil
}

// SendPingMessage sends a PING message to a given contact
func (network *Network) SendPingMessage(contact *Contact) {
	message := fmt.Sprintf("PING %s", network.LocalID.String())
	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending PING message:", err)
		return
	}
	fmt.Println("Received response:", response)
}
