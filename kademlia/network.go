package kademlia

import (
	"fmt"
	"net"
	"strings"
)

type Network struct {
	LocalID     *KademliaID
	LocalAddr   string
	pingPongCh  chan ChMsgs
	cmdChannel  chan CmdChMsgs  // <---
	dataChannel chan DataChMsgs // --->
}

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
		go network.parseConnection(conn)
	}
}

func (network *Network) parseConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	message := string(buffer[:n])
	message = strings.TrimSpace(message)

	switch {
	case strings.HasPrefix(message, "PING"):
		fmt.Println("Received NET message:", message)
		var pingOriginAddr, pingOriginID string
		_, err := fmt.Sscanf(message, "PING from %s %s", &pingOriginAddr, &pingOriginID)
		if err != nil {
			fmt.Println("Error parsing PING message:", err)
			return
		}

		network.pingPongCh <- ChMsgs{
			ChCmd:      "PING",
			SenderAddr: pingOriginAddr,
			SenderID:   pingOriginID,
		}

		pongMsgs := <-network.pingPongCh

		if pongMsgs.ChCmd == "PONG" {
			network.sendPongResponse(pingOriginAddr)
		} else {
			fmt.Println("Unexpected message received:", pongMsgs)
		}

	case strings.HasPrefix(message, "PONG"):
		fmt.Println("Received NET message:", message)
		var pongOriginAddr, pongOriginID string
		_, err := fmt.Sscanf(message, "PONG from %s %s", &pongOriginAddr, &pongOriginID)
		if err != nil {
			fmt.Println("Error parsing PONG message:", err)
			return
		}

	case strings.HasPrefix(message, "FIND_NODE"):
		fmt.Println("Parse con, FIND_NODE case")

		network.cmdChannel <- CmdChMsgs{
			KademCmd: "FIND_NODE",
		}

	default:
		fmt.Println("Received unknown message:", message)
	}
}

func (network *Network) SendPingMessage(remoteContact *Contact) {
	fmt.Println("Sending NET PING to remoteContact: ", remoteContact)
	conn, err := net.Dial("tcp", remoteContact.Address)
	if err != nil {
		fmt.Println("Error connecting to contact:", err)
		return
	}
	defer conn.Close()

	pingMessage := fmt.Sprintf("PING from %s %s", network.LocalAddr, network.LocalID.String())

	_, err = conn.Write([]byte(pingMessage))
	if err != nil {
		fmt.Println("Error sending ping message:", err)
	}
}

func (network *Network) sendPongResponse(pingOriginAddr string) {
	fmt.Println("Sending NET PONG to remoteContact: ", pingOriginAddr)
	pongIP := net.JoinHostPort(pingOriginAddr, "8080")

	conn, err := net.Dial("tcp", pongIP)
	if err != nil {
		fmt.Println("Error connecting back to origin on port 8080:", err)
		return
	}
	defer conn.Close()

	pongMessage := fmt.Sprintf("PONG from %s %s", network.LocalAddr, network.LocalID.String())
	_, err = conn.Write([]byte(pongMessage))
	if err != nil {
		fmt.Println("Error sending PONG message:", err)
		return
	}

	fmt.Println("Sent PONG message with server IP:", network.LocalAddr, "to", pongIP)
}

// SendFindNode sends a FIND_NODE message to a given contact
func (network *Network) SendFindNode(contact *Contact, targetID *KademliaID) ([]Contact, error) {
	message := fmt.Sprintf("FIND_NODE %s", targetID.String())
	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending FIND_NODE message:", err)
		return nil, err
	}

	// Handle the response
	if strings.HasPrefix(response, "FIND_NODE_RESPONSE") {
		contactsStr := strings.TrimPrefix(response, "FIND_NODE_RESPONSE ")
		contactsList := strings.Split(contactsStr, ";")
		var contacts []Contact
		for _, contactStr := range contactsList {
			parts := strings.Split(contactStr, "|")
			if len(parts) != 2 {
				continue
			}
			idStr := parts[0]
			address := parts[1]
			id := NewKademliaID(idStr)
			contact := NewContact(id, address)
			contacts = append(contacts, contact)
		}
		return contacts, nil
	} else {
		fmt.Println("Invalid FIND_NODE response:", response)
		return nil, fmt.Errorf("Invalid FIND_NODE response")
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
