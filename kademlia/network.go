package kademlia

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"
)

type Network struct {
	LocalID   *KademliaID
	LocalAddr string
	handler   MessageHandler
}

type MessageHandler interface {
	HandleMessage(message string, senderAddr string) string
}

// port for incoming traffic is set in Dockerfile and compose file
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

	remoteIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		fmt.Println("Error parsing remote address:", err)
		return
	}

	// all nodes are set to listen on 8080
	senderAddress := net.JoinHostPort(remoteIP, "8080")

	if network.handler != nil {
		response := network.handler.HandleMessage(message, senderAddress)
		if response != "" {
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error sending response:", err)
			}
		}
	} else {
		fmt.Println("No message handler set")
	}
}

// generic send msgs function to send all the messages commands of the network
func (network *Network) SendMessage(address string, message string) (string, error) {

	//check if port is missing add 8080
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%s:8080", address)
	}

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

func (network *Network) SendPing(contact *Contact) error {
	message := fmt.Sprintf("PING %s", network.LocalID.String())
	fmt.Println("Sending PING message to", contact.Address)
	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending PING message:", err)
		return err
	}
	if network.handler != nil {
		network.handler.HandleMessage(response, contact.Address)
	}
	return nil
}

func (network *Network) SendFindNode(contact *Contact, targetID *KademliaID) ([]Contact, error) {
	message := fmt.Sprintf("FIND_NODE %s", targetID.String())
	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending FIND_NODE message:", err)
		return nil, err
	}

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

func (network *Network) SendStore(contact *Contact, hash string, data []byte) error {
	dataBase64 := base64.StdEncoding.EncodeToString(data)
	message := fmt.Sprintf("STORE %s %s", hash, dataBase64)
	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending STORE message:", err)
		return err
	}
	if response != "STORE_OK" {
		fmt.Println("Error response from STORE:", response)
		return fmt.Errorf("Store failed: %s", response)
	}
	return nil
}

func (network *Network) SendFindValue(contact *Contact, key string) ([]byte, bool, error) {
	message := fmt.Sprintf("FIND_VALUE %s", key)

	response, err := network.SendMessage(contact.Address, message)
	if err != nil {
		return nil, false, err
	}

	if strings.HasPrefix(response, "VALUE ") {
		dataBase64 := strings.TrimPrefix(response, "VALUE ")
		data, err := base64.StdEncoding.DecodeString(dataBase64)
		if err != nil {
			return nil, false, err
		}
		return data, true, nil
	} else if response == "VALUE_NOT_FOUND" {
		return nil, false, nil
	} else {
		return nil, false, fmt.Errorf("unexpected response: %s", response)
	}
}
