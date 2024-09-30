package kademlia

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Kademlia represents the Kademlia distributed hash table.
type Kademlia struct {
	Network                   *Network
	RoutingTableActionChannel chan RoutingTableAction // The channel to handle routing table actions
	DataStoreActionChannel    chan DataStoreAction    // The channel to handle data store actions
	DataStore                 map[string][]byte
	NodeCli                   *NodeCli
}

// RoutingTableAction represents an action that needs to be performed on the RoutingTable.
type RoutingTableAction struct {
	ActionType string                    // Type of action (e.g., "AddContact", "FindClosestContacts")
	Contact    *Contact                  // Contact to add (for AddContact)
	TargetID   *KademliaID               // TargetID (for FindClosestContacts)
	ResponseCh chan RoutingTableResponse // Channel to send the response back to the caller
}

// RoutingTableResponse is a generic response for routing table actions.
// RoutingTableResponse is a generic response for routing table actions.
type RoutingTableResponse struct {
	ClosestContacts    []Contact // The closest contacts (for FindClosestContacts)
	routingTableString string    // The string representation of the routing table
}

// DataStoreAction represents an action to be performed on the DataStore.
type DataStoreAction struct {
	ActionType string                 // "Store" or "Retrieve"
	Key        string                 // The key (hash) for the data
	Value      []byte                 // The data to store (for "Store" action)
	ResponseCh chan DataStoreResponse // Channel to send the response
}

// DataStoreResponse represents the response from a DataStore action.
type DataStoreResponse struct {
	Value   []byte // The data retrieved (for "Retrieve" action)
	Success bool   // Indicates if the action was successful
}

// NewKademlia initializes a new Kademlia instance.
func NewKademlia(localAddr string) *Kademlia {

	network := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	kademlia := &Kademlia{
		Network:                   network,
		RoutingTableActionChannel: make(chan RoutingTableAction),
		DataStoreActionChannel:    make(chan DataStoreAction),
	}

	cli := &NodeCli{
		nodeCli: kademlia,  // Set the Kademlia instance as the CLI handler
	}

	kademlia.NodeCli = cli

	go cli.StartCLI()
	go kademlia.dataStoreWorker()
	go kademlia.routingTableWorker()

	network.handler = kademlia

	return kademlia
}


func (kademlia *Kademlia) JoinNetwork(root *Contact) {
	// Send a PING message to the contact
	kademlia.Network.SendPing(root)

	time.Sleep(1 * time.Second)

	kademlia.LookupContact(kademlia.Network.LocalID)
}

// routingTableWorker processes routing table actions sequentially.
func (kademlia *Kademlia) routingTableWorker() {
	routingTable := NewRoutingTable(NewContact(kademlia.Network.LocalID, kademlia.Network.LocalAddr))

	for action := range kademlia.RoutingTableActionChannel {
		switch action.ActionType {
		case "AddContact":
			routingTable.AddContact(*action.Contact)
			action.ResponseCh <- RoutingTableResponse{} // Send an empty response to acknowledge completion

		case "FindClosestContacts":
			closestContacts := routingTable.FindClosestContacts(action.TargetID, bucketSize)
			action.ResponseCh <- RoutingTableResponse{ClosestContacts: closestContacts}

		case "PrintRoutingTable":
			routingTableString := routingTable.PrintBuckets()
			action.ResponseCh <- RoutingTableResponse{routingTableString: routingTableString}

		}
	}
}

// dataStoreWorker processes data store actions sequentially.
func (kademlia *Kademlia) dataStoreWorker() {
	dataStore := make(map[string][]byte)

	for action := range kademlia.DataStoreActionChannel {
		switch action.ActionType {
		case "Store":
			dataStore[action.Key] = action.Value
			action.ResponseCh <- DataStoreResponse{Success: true}
		case "Retrieve":
			value, found := dataStore[action.Key]
			action.ResponseCh <- DataStoreResponse{Value: value, Success: found}
		}
	}
}

func (kademlia *Kademlia) PrintRoutingTable() {
	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "PrintRoutingTable",
		ResponseCh: responseCh,
	}

	response := <-responseCh                         // Receive the response from the channel
	fmt.Println("frog", response.routingTableString) // Print the routing table string
}

// HandleMessage implements the MessageHandler interface.
func (kademlia *Kademlia) HandleMessage(message string, senderAddr string) string {
	switch {
	case strings.HasPrefix(message, "PING"):
		kademlia.PrintRoutingTable()
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

func (kademlia *Kademlia) HandleCliMessage(message string, senderAddr string) string {
	parts := strings.Split(message, " ")
	switch parts[0] {
	case "put":
		if len(parts) < 2 {
			return "Usage: put <file_content (strings for now)>"
		}
		content := parts[1]
		err := kademlia.StoreData([]byte(content))
		if err != nil {
			return fmt.Sprintf("Error storing data: %v", err)
		}
		hash := kademlia.HashData([]byte(content))
		return fmt.Sprintf("Data stored with hash: %s", hash)
	case "get":
		if len(parts) < 2 {
			return "Usage: get <hash>"
		}
		hash := parts[1]
		data, err := kademlia.RetrieveData(hash)
		if err != nil {
			return fmt.Sprintf("Error retrieving data: %v", err)
		}
		return fmt.Sprintf("Data retrieved: %s", string(data))
	case "exit":
		return "Exiting CLI..."
	default:
		return "Unknown command"
	}
}

func (kademlia *Kademlia) handlePingMessage(message string, senderAddr string) string {
	fmt.Println("Received PING message from", senderAddr)
	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid PING message format")
		return ""
	}
	senderIDStr := parts[1]
	senderID := NewKademliaID(senderIDStr)
	senderContact := NewContact(senderID, senderAddr)

	// Create a response channel for this request
	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "AddContact",
		Contact:    &senderContact,
		ResponseCh: responseCh,
	}

	// Wait for acknowledgment
	<-responseCh

	// Prepare PONG response including our own ID
	response := fmt.Sprintf("PONG %s", kademlia.Network.LocalID.String())
	return response
}

func (kademlia *Kademlia) handlePongMessage(message string, senderAddr string) {
	fmt.Println("Received PONG message from", senderAddr)
	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid PONG message format")
		return
	}
	senderIDStr := parts[1]
	senderID := NewKademliaID(senderIDStr)
	senderContact := NewContact(senderID, senderAddr)

	// Create a response channel for this request
	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "AddContact",
		Contact:    &senderContact,
		ResponseCh: responseCh,
	}

	// Wait for acknowledgment
	<-responseCh
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

	// Create a response channel for this request
	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "FindClosestContacts",
		TargetID:   targetID,
		ResponseCh: responseCh,
	}

	// Wait for the result
	response := <-responseCh
	closestContacts := response.ClosestContacts

	// Prepare response
	var contactsStrList []string
	for _, contact := range closestContacts {
		contactStr := fmt.Sprintf("%s|%s", contact.ID.String(), contact.Address)
		contactsStrList = append(contactsStrList, contactStr)
	}
	responseMessage := "FIND_NODE_RESPONSE " + strings.Join(contactsStrList, ";")

	return responseMessage
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

	// Initialize the shortlist with k closest contacts from the routing table
	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "FindClosestContacts",
		TargetID:   targetID,
		ResponseCh: responseCh,
	}
	response := <-responseCh
	closestContacts := response.ClosestContacts

	shortlist := make(map[string]Contact)
	queried := make(map[string]bool)

	for _, contact := range closestContacts {
		shortlist[contact.ID.String()] = contact
	}

	for {
		// Find the alpha closest unqueried contacts
		var unqueriedContacts []Contact
		for _, contact := range shortlist {
			if !queried[contact.ID.String()] {
				unqueriedContacts = append(unqueriedContacts, contact)
			}
		}

		if len(unqueriedContacts) == 0 {
			break
		}

		// Sort unqueriedContacts by distance to the target ID
		sort.Slice(unqueriedContacts, func(i, j int) bool {
			return unqueriedContacts[i].ID.CalcDistance(targetID).Less(unqueriedContacts[j].ID.CalcDistance(targetID))
		})

		// Select α closest contacts to query in parallel
		closestAlpha := unqueriedContacts
		if len(unqueriedContacts) > alpha {
			closestAlpha = unqueriedContacts[:alpha]
		}

		// Use WaitGroup to wait for parallel queries to complete
		var wg sync.WaitGroup
		resultsChannel := make(chan []Contact, len(closestAlpha))

		for _, contact := range closestAlpha {
			queried[contact.ID.String()] = true
			wg.Add(1)

			go func(contact Contact) {

				defer wg.Done()
				contactsReceived := kademlia.SendFindNode(targetID, &contact)

				resultsChannel <- contactsReceived

			}(contact)

			kademlia.RoutingTableActionChannel <- RoutingTableAction{
				ActionType: "AddContact",
				Contact:    &contact,
				ResponseCh: responseCh,
			}
			<-responseCh
		}
		wg.Wait()
		close(resultsChannel)

		// Collect results from all parallel queries
		for contactsReceived := range resultsChannel {
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

	// Convert the shortlist to a slice and sort it by distance to the target ID
	var finalClosest []Contact
	for _, contact := range shortlist {
		finalClosest = append(finalClosest, contact)
	}

	sort.Slice(finalClosest, func(i, j int) bool {
		return finalClosest[i].ID.CalcDistance(targetID).Less(finalClosest[j].ID.CalcDistance(targetID))
	})

	// Return the k closest contacts
	if len(finalClosest) > k {
		finalClosest = finalClosest[:k]
	}

	return finalClosest
}

// STORE LOGIC

// SendStore sends a STORE message to the given contact to store the data.
func (kademlia *Kademlia) SendStore(contact Contact, key string, data []byte) error {
	return kademlia.Network.SendStore(&contact, key, data)
}

// StoreData stores the given data in the network according to Kademlia protocol.
func (kademlia *Kademlia) StoreData(data []byte) error {
	// Step 1: Hash the data to obtain the key
	key := kademlia.HashData(data)
	targetID := NewKademliaID(key)

	// Step 2: Perform a node lookup for the key
	closestContacts := kademlia.LookupContact(targetID)

	if len(closestContacts) == 0 {
		return fmt.Errorf("no contacts found to store the data")
	}

	// Step 3: Send STORE messages to the k closest nodes
	var wg sync.WaitGroup
	for _, contact := range closestContacts {
		wg.Add(1)
		go func(contact Contact) {
			defer wg.Done()
			err := kademlia.SendStore(contact, key, data)
			if err != nil {
				fmt.Printf("Error storing data on %s: %v\n", contact.Address, err)
			}
		}(contact)
	}

	// Wait for all STORE RPCs to complete
	wg.Wait()
	return nil
}

// RetrieveData retrieves data associated with the given key from the network.
func (kademlia *Kademlia) RetrieveData(key string) ([]byte, error) {
	targetID := NewKademliaID(key)

	// Step 1: Perform a node lookup for the key
	closestContacts := kademlia.LookupContact(targetID)

	if len(closestContacts) == 0 {
		return nil, fmt.Errorf("no contacts found to retrieve the data")
	}

	// Step 2: Send FIND_VALUE messages to the closest contacts
	for _, contact := range closestContacts {
		data, found, err := kademlia.SendFindValue(contact, key)
		if err != nil {
			fmt.Printf("Error retrieving data from %s: %v\n", contact.Address, err)
			continue
		}
		if found {
			return data, nil
		}
	}

	// If data not found, return an error
	return nil, fmt.Errorf("data not found in the network")
}

// SendFindValue sends a FIND_VALUE message to the given contact.
func (kademlia *Kademlia) SendFindValue(contact Contact, key string) ([]byte, bool, error) {
	data, found, err := kademlia.Network.SendFindValue(&contact, key)
	if err != nil {
		return nil, false, err
	}
	return data, found, nil
}

// HashData computes the SHA1 hash of the given data.
func (kademlia *Kademlia) HashData(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
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

	// Use DataStoreActionChannel to store data
	responseCh := make(chan DataStoreResponse)
	kademlia.DataStoreActionChannel <- DataStoreAction{
		ActionType: "Store",
		Key:        hash,
		Value:      data,
		ResponseCh: responseCh,
	}

	// Wait for the response
	response := <-responseCh
	if response.Success {
		fmt.Println("Stored data with hash:", hash)
		return "STORE_OK"
	} else {
		return "STORE_ERROR Failed to store data"
	}
}

func (kademlia *Kademlia) handleFindValueMessage(message string, senderAddr string) string {
	fmt.Println("Received FIND_VALUE message from", senderAddr)

	parts := strings.Split(message, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid FIND_VALUE message format")
		return ""
	}
	hash := parts[1]

	// Use DataStoreActionChannel to retrieve data
	responseCh := make(chan DataStoreResponse)
	kademlia.DataStoreActionChannel <- DataStoreAction{
		ActionType: "Retrieve",
		Key:        hash,
		ResponseCh: responseCh,
	}

	// Wait for the response
	response := <-responseCh
	if response.Success {
		responseMessage := fmt.Sprintf("VALUE %s", base64.StdEncoding.EncodeToString(response.Value))
		return responseMessage
	} else {
		responseMessage := "VALUE_NOT_FOUND"
		return responseMessage
	}
}
