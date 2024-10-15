package kademlia

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Kademlia struct {
	Network                   *Network
	RoutingTableActionChannel chan RoutingTableAction // channel to handle routing table actions
	DataStoreActionChannel    chan DataStoreAction    // channel to handle data store actions
	DataStore                 map[string][]byte
	NodeCli                   *NodeCli
}

type RoutingTableAction struct {
	ActionType string
	Contact    *Contact
	TargetID   *KademliaID
	ResponseCh chan RoutingTableResponse
}

type RoutingTableResponse struct {
	ClosestContacts    []Contact
	routingTableString string
}

type DataStoreAction struct {
	ActionType string
	Key        string
	Value      []byte
	ResponseCh chan DataStoreResponse
}

type DataStoreResponse struct {
	Value   []byte
	Success bool
}

// init func for kademlia instance
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
		nodeCli: kademlia, // set the Kademlia instance as the CLI handler
	}

	kademlia.NodeCli = cli

	go cli.StartCLI() // listens for input form terminal, run "docker attach psID"
	go kademlia.dataStoreWorker()
	go kademlia.routingTableWorker()

	network.handler = kademlia // in go the interface is implemented like this since kademlia has the HandleMessages func

	return kademlia
}

func (kademlia *Kademlia) JoinNetwork(root *Contact) {
	kademlia.Network.SendPing(root)
	time.Sleep(1 * time.Second)
	kademlia.LookupContact(kademlia.Network.LocalID) // lookup on self to populate rt
}

func (kademlia *Kademlia) routingTableWorker() {
	routingTable := NewRoutingTable(NewContact(kademlia.Network.LocalID, kademlia.Network.LocalAddr))

	for action := range kademlia.RoutingTableActionChannel {
		switch action.ActionType {
		case "AddContact":
			routingTable.AddContact(*action.Contact)
			action.ResponseCh <- RoutingTableResponse{}

		case "FindClosestContacts":
			closestContacts := routingTable.FindClosestContacts(action.TargetID, bucketSize)
			action.ResponseCh <- RoutingTableResponse{ClosestContacts: closestContacts}

		case "PrintRoutingTable":
			routingTableString := routingTable.PrintBuckets()
			action.ResponseCh <- RoutingTableResponse{routingTableString: routingTableString}
		}
	}
}

func (kademlia *Kademlia) dataStoreWorker() {
	dataStorePath := "./dataStore/"

	if _, err := os.Stat(dataStorePath); os.IsNotExist(err) {
		err := os.Mkdir(dataStorePath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating data store directory:", err)
			return
		}
	}

	for action := range kademlia.DataStoreActionChannel {
		switch action.ActionType {

		case "Store":
			filePath := dataStorePath + action.Key
			err := ioutil.WriteFile(filePath, action.Value, 0644)
			if err != nil {
				action.ResponseCh <- DataStoreResponse{Success: false}
				fmt.Println("Error writing to file:", err)
			} else {
				action.ResponseCh <- DataStoreResponse{Success: true}
			}

		case "Retrieve":
			filePath := dataStorePath + action.Key
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					action.ResponseCh <- DataStoreResponse{Value: nil, Success: false}
				} else {
					action.ResponseCh <- DataStoreResponse{Value: nil, Success: false}
					fmt.Println("Error reading from file:", err)
				}
			} else {
				action.ResponseCh <- DataStoreResponse{Value: data, Success: true}
			}
		}
	}
}

// helper func for dev cycle
func (kademlia *Kademlia) PrintRoutingTable() {
	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "PrintRoutingTable",
		ResponseCh: responseCh,
	}

	response := <-responseCh // Receive the response from the channel
	fmt.Println(response.routingTableString)
}

// HandleMessage implements the MessageHandler interface in the Network module
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

	responseCh := make(chan RoutingTableResponse)
	kademlia.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "AddContact",
		Contact:    &senderContact,
		ResponseCh: responseCh,
	}

	// Wait for acknowledgment
	<-responseCh

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

func (kademlia *Kademlia) Ping(contact *Contact) {
	err := kademlia.Network.SendPing(contact)
	if err != nil {
		fmt.Println("Error pinging contact:", err)
	}
}

func (kademlia *Kademlia) SendFindNode(targetID *KademliaID, contact *Contact) []Contact {
	contacts, err := kademlia.Network.SendFindNode(contact, targetID)
	if err != nil {
		fmt.Println("Error during SendFindNode:", err)
		return nil
	}
	return contacts
}

func (kademlia *Kademlia) LookupContact(targetID *KademliaID) []Contact {
	alpha := 3
	k := bucketSize

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
		var unqueriedContacts []Contact
		for _, contact := range shortlist {
			if !queried[contact.ID.String()] {
				unqueriedContacts = append(unqueriedContacts, contact)
			}
		}

		if len(unqueriedContacts) == 0 {
			break
		}

		sort.Slice(unqueriedContacts, func(i, j int) bool {
			return unqueriedContacts[i].ID.CalcDistance(targetID).Less(unqueriedContacts[j].ID.CalcDistance(targetID))
		})

		closestAlpha := unqueriedContacts
		if len(unqueriedContacts) > alpha {
			closestAlpha = unqueriedContacts[:alpha]
		}

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

	var finalClosest []Contact
	for _, contact := range shortlist {
		finalClosest = append(finalClosest, contact)
	}

	sort.Slice(finalClosest, func(i, j int) bool {
		return finalClosest[i].ID.CalcDistance(targetID).Less(finalClosest[j].ID.CalcDistance(targetID))
	})

	if len(finalClosest) > k {
		finalClosest = finalClosest[:k]
	}

	return finalClosest
}

func (kademlia *Kademlia) SendStore(contact Contact, key string, data []byte) error {
	return kademlia.Network.SendStore(&contact, key, data)
}

func (kademlia *Kademlia) StoreData(data []byte) error {
	key := kademlia.HashData(data)
	targetID := NewKademliaID(key)

	closestContacts := kademlia.LookupContact(targetID)

	if len(closestContacts) == 0 {
		return fmt.Errorf("no contacts found to store the data")
	}

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

	wg.Wait()
	return nil
}

func (kademlia *Kademlia) RetrieveData(key string) ([]byte, error) {
	targetID := NewKademliaID(key)

	closestContacts := kademlia.LookupContact(targetID)

	if len(closestContacts) == 0 {
		return nil, fmt.Errorf("no contacts found to retrieve the data")
	}

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

	return nil, fmt.Errorf("data not found in the network")
}

func (kademlia *Kademlia) SendFindValue(contact Contact, key string) ([]byte, bool, error) {
	data, found, err := kademlia.Network.SendFindValue(&contact, key)
	if err != nil {
		return nil, false, err
	}
	return data, found, nil
}

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
