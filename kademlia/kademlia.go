package kademlia

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
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
func (kademlia *Kademlia) HandleMessage(message string, senderAddr string) string {
	switch {
	case strings.HasPrefix(message, "PING"):
		return kademlia.handlePingMessage(message, senderAddr)
	case strings.HasPrefix(message, "PONG"):
		kademlia.handlePongMessage(message, senderAddr)
		return ""
	case strings.HasPrefix(message, "FIND_NODE"):
		return kademlia.handleFindNodeMessage(message, senderAddr)
	case strings.HasPrefix(message, "FIND_VALUE"):
		return kademlia.handleFindValueMessage(message, senderAddr)
	case strings.HasPrefix(message, "STORE"):
		return kademlia.handleStoreMessage(message, senderAddr)
	default:
		fmt.Println("Received unknown message:", message)
		return ""
	}
}

func (kademlia *Kademlia) handlePingMessage(message string, senderAddr string) string {
	fmt.Println("Received PING message from", senderAddr)
	// Extract sender's ID from the message
	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid PING message format")
		return ""
	}
	senderIDStr := parts[1]
	senderID := NewKademliaID(senderIDStr)
	senderContact := NewContact(senderID, senderAddr)
	kademlia.RoutingTable.AddContact(senderContact)

	// Prepare PONG response including our own ID
	response := fmt.Sprintf("PONG %s", kademlia.Network.LocalID.String())
	return response
}

func (kademlia *Kademlia) handlePongMessage(message string, senderAddr string) {
	fmt.Println("Received PONG message from", senderAddr)
	// Extract sender's ID from the message
	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid PONG message format")
		return
	}
	senderIDStr := parts[1]
	senderID := NewKademliaID(senderIDStr)
	senderContact := NewContact(senderID, senderAddr)
	kademlia.RoutingTable.AddContact(senderContact)
}

func (kademlia *Kademlia) handleFindNodeMessage(message string, senderAddr string) string {
	fmt.Println("Received FIND_NODE message from", senderAddr)
	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid FIND_NODE message format")
		return ""
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
	return responseMessage
}

func (kademlia *Kademlia) handleStoreMessage(message string, senderAddr string) string {
	fmt.Println("Received STORE message from", senderAddr)

	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		fmt.Println("Invalid STORE message format")
		return "STORE_ERROR Invalid format"
	}
	hash := parts[1]
	dataBase64 := parts[2]
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		fmt.Println("Error decoding data:", err)
		return "STORE_ERROR Decoding error"
	}

	kademlia.DataStore[hash] = data
	fmt.Println("Stored data with hash:", hash)

	// Return acknowledgment
	return "STORE_OK"
}

func (kademlia *Kademlia) handleFindValueMessage(message string, senderAddr string) string {
	fmt.Println("Received FIND_VALUE message from", senderAddr)

	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid FIND_VALUE message format")
		return ""
	}
	hash := parts[1]

	if data, found := kademlia.DataStore[hash]; found {
		responseMessage := fmt.Sprintf("VALUE %s", base64.StdEncoding.EncodeToString(data))
		return responseMessage
	} else {
		// Optionally, send a response indicating data not found
		responseMessage := "VALUE_NOT_FOUND"
		return responseMessage
	}
}

// Ping a contact
func (kademlia *Kademlia) Ping(contact *Contact) {
	err := kademlia.Network.SendPing(contact)
	if err != nil {
		fmt.Println("Error pinging contact:", err)
	}
}

// SendFindNode sends a FIND_NODE message to a contact
func (kademlia *Kademlia) SendFindNode(targetID *KademliaID, contact *Contact) []Contact {
	contacts, err := kademlia.Network.SendFindNode(contact, targetID)
	if err != nil {
		fmt.Println("Error during SendFindNode:", err)
		return nil
	}
	return contacts
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
	fmt.Println("Storing data with hash:", hashString)

	// Find the k closest nodes to the hash
	targetID := NewKademliaID(hashString)
	closestContacts := kademlia.LookupContact(targetID)
	fmt.Println("Closest contacts to hash:", closestContacts)

	// Send STORE messages to each contact
	for _, contact := range closestContacts {
		kademlia.Network.SendStore(&contact, hashString, data)
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

// LookupData retrieves data from the network
func (kademlia *Kademlia) LookupData(hash string) ([]byte, error) {
	// Check local datastore
	if data, found := kademlia.DataStore[hash]; found {
		return data, nil
	}

	// Perform iterative lookup
	targetID := NewKademliaID(hash)
	closestContacts := kademlia.LookupContact(targetID)

	for _, contact := range closestContacts {
		data, err := kademlia.Network.SendFindValue(&contact, hash)
		if err == nil && data != nil {
			// Store data locally
			kademlia.DataStore[hash] = data
			return data, nil
		}
	}

	return nil, fmt.Errorf("Data not found for hash: %s", hash)
}
