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
	DataStore                 map[string][]byte
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

// NewKademlia initializes a new Kademlia instance.
func NewKademlia(localAddr string) *Kademlia {

	network := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	if localAddr == "172.16.238.10" {
		network = &Network{
			LocalID:   NewKademliaID("FFFFFFFF00000000000000000000000000000000"),
			LocalAddr: localAddr,
		}
	}

	// Channel to process routing table actions
	routingTableActionChannel := make(chan RoutingTableAction)

	kademlia := &Kademlia{
		Network:                   network,
		RoutingTableActionChannel: routingTableActionChannel,
		DataStore:                 make(map[string][]byte),
	}

	// Start the routing table worker
	go kademlia.routingTableWorker()

	// Set the network's message handler to this Kademlia instance
	network.handler = kademlia

	return kademlia
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

				time.Sleep(3 * time.Second)
				fmt.Println("PARALLELLT PLS")

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

// HashData computes the SHA1 hash of the given data.
func (kademlia *Kademlia) HashData(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}
