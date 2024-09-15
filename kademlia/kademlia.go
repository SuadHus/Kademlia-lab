package kademlia

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strings"
)

// Kademlia represents the Kademlia distributed hash table.
type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
	DataStore    map[string][]byte
}

// NewKademlia initializes a new Kademlia instance.
func NewKademlia(localAddr string) *Kademlia {

	network := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	routingTable := NewRoutingTable(NewContact(network.LocalID, localAddr))

	kademlia := &Kademlia{
		Network:      network,
		RoutingTable: routingTable,
		DataStore:    make(map[string][]byte),
	}

	// Set the network's message handler to this Kademlia instance
	network.handler = kademlia

	return kademlia
}

// HandleMessage implements the MessageHandler interface
func (kademlia *Kademlia) HandleMessage(message string, conn net.Conn) {
	switch {
	case strings.HasPrefix(message, "PING"):
		kademlia.handlePingMessage(conn, message)
	case strings.HasPrefix(message, "PONG"):
		// Handle PONG if needed
	case strings.HasPrefix(message, "FIND_NODE"):
		kademlia.handleFindNodeMessage(conn, message)
	case strings.HasPrefix(message, "FIND_VALUE"):
		kademlia.handleFindValueMessage(conn, message)
	case strings.HasPrefix(message, "STORE"):
		kademlia.handleStoreMessage(conn, message)
	default:
		fmt.Println("Received unknown message:", message)
	}
}

func (kademlia *Kademlia) handlePingMessage(conn net.Conn, message string) {
	fmt.Println("Received PING message from", conn.RemoteAddr())
	// Extract sender ID
	senderIDStr := strings.TrimSpace(strings.TrimPrefix(message, "PING"))
	senderID := NewKademliaID(senderIDStr)
	senderAddr := conn.RemoteAddr().String()
	senderContact := NewContact(senderID, senderAddr)
	kademlia.RoutingTable.AddContact(senderContact)

	// Send PONG response
	response := fmt.Sprintf("PONG %s", kademlia.Network.LocalID.String())
	conn.Write([]byte(response))
}

func (kademlia *Kademlia) handleFindNodeMessage(conn net.Conn, message string) {
	fmt.Println("Received FIND_NODE message from", conn.RemoteAddr())
	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid FIND_NODE message format")
		return
	}
	targetIDStr := parts[1]
	targetID := NewKademliaID(targetIDStr)

	// Find closest contacts
	closestContacts := kademlia.RoutingTable.FindClosestContacts(targetID, bucketSize)

	// Prepare response
	var contactsStrList []string
	for _, contact := range closestContacts {
		contactStr := fmt.Sprintf("%s|%s", contact.ID.String(), contact.Address)
		contactsStrList = append(contactsStrList, contactStr)
	}
	responseMessage := "FIND_NODE_RESPONSE " + strings.Join(contactsStrList, ";")

	// Send response
	conn.Write([]byte(responseMessage))
}

func (kademlia *Kademlia) handleStoreMessage(conn net.Conn, message string) {
	fmt.Println("Received STORE message from", conn.RemoteAddr())

	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		fmt.Println("Invalid STORE message format")
		return
	}
	hash := parts[1]
	dataBase64 := parts[2]
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		fmt.Println("Error decoding data:", err)
		return
	}

	kademlia.DataStore[hash] = data
	fmt.Println("Stored data with hash:", hash)
}

func (kademlia *Kademlia) handleFindValueMessage(conn net.Conn, message string) {
	fmt.Println("Received FIND_VALUE message from", conn.RemoteAddr())

	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid FIND_VALUE message format")
		return
	}
	hash := parts[1]

	if data, found := kademlia.DataStore[hash]; found {
		responseMessage := fmt.Sprintf("VALUE %s", base64.StdEncoding.EncodeToString(data))
		conn.Write([]byte(responseMessage))
		fmt.Println("Sent VALUE response to", conn.RemoteAddr())
	} else {
		// Optionally, send a response indicating data not found
		responseMessage := "VALUE_NOT_FOUND"
		conn.Write([]byte(responseMessage))
	}
}

// SendFindNode sends a FIND_NODE message to a contact
func (kademlia *Kademlia) SendFindNode(targetID *KademliaID, contact *Contact) []Contact {
	message := fmt.Sprintf("FIND_NODE %s", targetID.String())
	response, err := kademlia.Network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending FIND_NODE:", err)
		return nil
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
		return contacts
	} else {
		fmt.Println("Invalid FIND_NODE response:", response)
		return nil
	}
}

// LookupContact performs an iterative node lookup
func (kademlia *Kademlia) LookupContact(targetID *KademliaID) []Contact {
	alpha := 3
	k := bucketSize

	// Initialize the shortlist with k closest contacts
	closestContacts := kademlia.RoutingTable.FindClosestContacts(targetID, k)
	shortlist := make(map[string]Contact)
	queried := make(map[string]bool)

	for _, contact := range closestContacts {
		shortlist[contact.ID.String()] = contact
	}

	for {
		// Find the α closest unqueried contacts
		var unqueriedContacts []Contact
		for _, contact := range shortlist {
			if !queried[contact.ID.String()] {
				unqueriedContacts = append(unqueriedContacts, contact)
			}
		}

		if len(unqueriedContacts) == 0 {
			break
		}

		// Sort unqueriedContacts by distance to target
		sort.Slice(unqueriedContacts, func(i, j int) bool {
			distI := unqueriedContacts[i].ID.CalcDistance(targetID)
			distJ := unqueriedContacts[j].ID.CalcDistance(targetID)
			return distI.Less(distJ)
		})

		// Select up to α contacts to query
		numToQuery := alpha
		if len(unqueriedContacts) < alpha {
			numToQuery = len(unqueriedContacts)
		}
		contactsToQuery := unqueriedContacts[:numToQuery]

		// For each contact to query
		for _, contact := range contactsToQuery {
			queried[contact.ID.String()] = true

			// Send FIND_NODE to the contact
			contactsReceived := kademlia.SendFindNode(targetID, &contact)

			// Add the contact to the routing table
			kademlia.RoutingTable.AddContact(contact)

			// Add any new contacts received to the shortlist
			for _, newContact := range contactsReceived {
				if newContact.ID.Equals(kademlia.Network.LocalID) {
					continue
				}
				if _, exists := shortlist[newContact.ID.String()]; !exists {
					shortlist[newContact.ID.String()] = newContact
				}
			}
		}
	}

	// Convert shortlist to a slice and sort
	var finalContacts []Contact
	for _, contact := range shortlist {
		finalContacts = append(finalContacts, contact)
	}

	// Sort the finalContacts by distance to target
	sort.Slice(finalContacts, func(i, j int) bool {
		distI := finalContacts[i].ID.CalcDistance(targetID)
		distJ := finalContacts[j].ID.CalcDistance(targetID)
		return distI.Less(distJ)
	})

	// Return the k closest contacts
	if len(finalContacts) > k {
		finalContacts = finalContacts[:k]
	}

	return finalContacts
}

// Store stores data in the network
func (kademlia *Kademlia) Store(data []byte) string {
	// Compute the hash of the data
	hash := sha1.Sum(data)
	hashString := hex.EncodeToString(hash[:])

	// Find the k closest nodes to the hash
	targetID := NewKademliaID(hashString)
	closestContacts := kademlia.LookupContact(targetID)

	// Send STORE messages to each contact
	for _, contact := range closestContacts {
		kademlia.SendStore(&contact, hashString, data)
	}

	// Optionally store the data locally if this node is among the closest
	for _, contact := range closestContacts {
		if contact.ID.Equals(kademlia.Network.LocalID) {
			kademlia.DataStore[hashString] = data
			break
		}
	}

	return hashString
}

// SendStore sends a STORE message to a contact
func (kademlia *Kademlia) SendStore(contact *Contact, hash string, data []byte) {
	dataBase64 := base64.StdEncoding.EncodeToString(data)
	message := fmt.Sprintf("STORE %s %s", hash, dataBase64)
	_, err := kademlia.Network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending STORE message:", err)
	}
}

// LookupData retrieves data from the network
func (kademlia *Kademlia) LookupData(hash string) ([]byte, error) {
	// Check local datastore
	if data, found := kademlia.DataStore[hash]; found {
		return data, nil
	}

	// Perform iterative lookup
	targetID := NewKademliaID(hash)
	// alpha := 3
	// k := bucketSize

	// Initialize shortlist
	closestContacts := kademlia.LookupContact(targetID)

	for _, contact := range closestContacts {
		data := kademlia.SendFindData(&contact, hash)
		if data != nil {
			// Store data locally
			kademlia.DataStore[hash] = data
			return data, nil
		}
	}

	return nil, fmt.Errorf("Data not found for hash: %s", hash)
}

// SendFindData sends a FIND_VALUE message to a contact
func (kademlia *Kademlia) SendFindData(contact *Contact, hash string) []byte {
	message := fmt.Sprintf("FIND_VALUE %s", hash)
	response, err := kademlia.Network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending FIND_VALUE:", err)
		return nil
	}

	if strings.HasPrefix(response, "VALUE") {
		dataBase64 := strings.TrimPrefix(response, "VALUE ")
		data, err := base64.StdEncoding.DecodeString(dataBase64)
		if err != nil {
			fmt.Println("Error decoding data:", err)
			return nil
		}
		return data
	}
	return nil
}

// SendPing sends a PING message to a contact
func (kademlia *Kademlia) SendPing(contact *Contact) {
	message := fmt.Sprintf("PING %s", kademlia.Network.LocalID.String())
	response, err := kademlia.Network.SendMessage(contact.Address, message)
	if err != nil {
		fmt.Println("Error sending PING message:", err)
		return
	}
	fmt.Println("Received response:", response)
}
