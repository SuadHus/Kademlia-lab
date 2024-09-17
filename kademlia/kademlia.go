package kademlia

import (
	"fmt"
)

// Kademlia represents the Kademlia distributed hash table.
type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
}

// NewKademlia initializes a new Kademlia instance.
func NewKademlia(localAddr string) *Kademlia {
	// Create a communication channel
	msgChan := make(chan string)

	// Declare the 'me' variable as a Contact
	var me Contact

	// Check local address and assign the correct Kademlia ID

	if localAddr == "172.16.238.10" {
		fmt.Println("Local address: ", localAddr)
		me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), localAddr)
	} else {
		me = NewContact(NewRandomKademliaID(), localAddr)
	}

	// Initialize the network with the channel
	network := &Network{
		LocalID:   me.ID, // Use 'me.ID' instead of NewRandomKademliaID to ensure consistency
		LocalAddr: localAddr,
		msgChan:   msgChan, // Pass the channel to the network
	}

	// Initialize the routing table with the 'me' contact
	routingTable := NewRoutingTable(me)
	fmt.Println("New Kademlia instance created", routingTable.me.ID)

	// Create Kademlia instance
	k := &Kademlia{
		Network:      network,
		RoutingTable: routingTable,
	}

	// Start listening for messages in a separate goroutine
	go k.listenForMessages()
	go network.Listen(localAddr, 8080)

	return k
}

func (kademlia *Kademlia) listenForMessages() {
	for {
		select {
		case msg := <-kademlia.Network.msgChan:
			// Process the received message from the channel

			fmt.Println("Received message on Kademlia channel:", msg)
		}
	}
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO: Implement contact lookup
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO: Implement data lookup
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO: Implement data storage
}
