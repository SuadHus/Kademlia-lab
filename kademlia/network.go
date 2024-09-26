package kademlia

import (
	//"encoding/base64"
	"fmt"
	"net"
	"strings"
)

// Network struct definition
type Network struct {
	LocalID   *KademliaID // The local node's Kademlia ID
	LocalAddr string      // The local node's address
	handler   MessageHandler
}

// MessageHandler interface for handling messages
type MessageHandler interface {
	HandleMessage(message string, senderAddr string) string
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

	fmt.Println("Listening on", addr)

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

	// Read message from connection
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	message := string(buffer[:n])
	message = strings.TrimSpace(message)

	// Get the remote IP address without the port
	remoteIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		fmt.Println("Error parsing remote address:", err)
		return
	}

	// Since all nodes listen on port 8080, append it to the IP
	senderAddress := net.JoinHostPort(remoteIP, "8080")

	// Pass the message and remote address to the handler and get a response
	if network.handler != nil {
		response := network.handler.HandleMessage(message, senderAddress)
		if response != "" {
			// Send response back
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error sending response:", err)
			}
		}
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

// SendPing sends a PING message to a given contact
func (network *Network) SendPing(contact *Contact) error {
	message := fmt.Sprintf("PING %s", network.LocalID.String())
	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending PING message:", err)
		return err
	}
	// Handle the response
	if network.handler != nil {
		network.handler.HandleMessage(response, contact.Address)
	}
	return nil
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

// func (network *Network) SendStore(contact *Contact, hash string, data []byte) error {
// 	dataBase64 := base64.StdEncoding.EncodeToString(data)
// 	message := fmt.Sprintf("STORE %s %s", hash, dataBase64)
// 	response, err := network.SendMessage(contact.Address, message)
// 	if err != nil {
// 		fmt.Println("Error sending STORE message:", err)
// 		return err
// 	}
// 	if response != "STORE_OK" {
// 		fmt.Println("Error response from STORE:", response)
// 		return fmt.Errorf("Store failed: %s", response)
// 	}
// 	return nil
// }

// // SendFindValue sends a FIND_VALUE message to a contact
// func (network *Network) SendFindValue(contact *Contact, hash string) ([]byte, error) {
// 	message := fmt.Sprintf("FIND_VALUE %s", hash)
// 	response, err := network.SendMessage(contact.Address, message)
// 	if err != nil {
// 		fmt.Println("Error sending FIND_VALUE message:", err)
// 		return nil, err
// 	}

// 	if strings.HasPrefix(response, "VALUE") {
// 		dataBase64 := strings.TrimPrefix(response, "VALUE ")
// 		data, err := base64.StdEncoding.DecodeString(dataBase64)
// 		if err != nil {
// 			fmt.Println("Error decoding data:", err)
// 			return nil, err
// 		}
// 		return data, nil
// 	} else {
// 		// Handle other responses
// 		return nil, fmt.Errorf("Invalid FIND_VALUE response")
// 	}
// }
