package kademlia

import (
	"fmt"
	"net"
	"strings"
)

type Network struct {
	LocalID     *KademliaID
	LocalAddr   string
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

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	message := string(buffer[:n])
	message = strings.TrimSpace(message)
	fmt.Println("Received NET message:", message)

	fields := strings.Fields(message)
	if len(fields) < 4 || fields[1] != "from" {
		fmt.Println("Invalid message format:", message)
		return
	}

	command := fields[0]
	senderAddr := fields[2]
	senderID := fields[3]
	var targetID string
	if len(fields) >= 5 {
		targetID = fields[4]
	}

	switch command {
	case "PING":
		fmt.Println("Received PING message from:", senderAddr)
		network.cmdChannel <- CmdChMsgs{
			KademCmd:   "PING",
			SenderAddr: senderAddr,
			SenderID:   senderID,
		}

		// Send PONG response over the same connection
		pongMessage := fmt.Sprintf("PONG from %s %s", network.LocalAddr, network.LocalID.String())
		_, err = conn.Write([]byte(pongMessage))
		if err != nil {
			fmt.Println("Error sending PONG response:", err)
			return
		}

	case "FIND_NODE":
		if targetID == "" {
			fmt.Println("Missing target ID in FIND_NODE message")
			return
		}
		fmt.Println("Received FIND_NODE message from:", senderAddr)
		network.cmdChannel <- CmdChMsgs{
			KademCmd:   "FIND_NODE",
			SenderAddr: senderAddr,
			SenderID:   senderID,
			TargetID:   targetID,
		}

		kademliaDataResponse := <-network.dataChannel

		// Send FIND_NODE_RESPONSE back over the same connection
		contactsStr := ""
		for _, contact := range kademliaDataResponse.Contacts {
			contactsStr += fmt.Sprintf("%s|%s;", contact.ID.String(), contact.Address)
		}

		response := fmt.Sprintf("FIND_NODE_RESPONSE %s", contactsStr)
		fmt.Println("Sending FIND_NODE_RESPONSE:", response)

		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error sending FIND_NODE_RESPONSE message:", err)
			return
		}

	default:
		fmt.Println("Received unknown command:", command)
	}
}

func (network *Network) SendPingMessage(remoteContact *Contact) {
	fmt.Println("Sending PING to remoteContact:", remoteContact)

	pingMessage := fmt.Sprintf("PING from %s %s", network.LocalAddr, network.LocalID.String())

	response, err := network.SendMessage(remoteContact.Address, pingMessage)
	if err != nil {
		fmt.Println("Error sending PING message:", err)
		return
	}

	fmt.Println("Received response to PING:", response)

	if strings.HasPrefix(response, "PONG") {
		var pongOriginAddr, pongOriginID string
		_, err := fmt.Sscanf(response, "PONG from %s %s", &pongOriginAddr, &pongOriginID)
		if err != nil {
			fmt.Println("Error parsing PONG response:", err)
			return
		}
		network.cmdChannel <- CmdChMsgs{
			KademCmd:   "PONG",
			SenderAddr: pongOriginAddr,
			SenderID:   pongOriginID,
		}
	} else {
		fmt.Println("Unexpected response to PING:", response)
	}
}

func (network *Network) SendFindNode(contact *Contact, targetID *KademliaID) ([]Contact, error) {
	message := fmt.Sprintf("FIND_NODE from %s %s %s", network.LocalAddr, network.LocalID.String(), targetID.String())
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
	conn, err := net.Dial("tcp", address+":8080")
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
