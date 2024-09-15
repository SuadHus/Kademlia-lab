package kademlia

import (
	"fmt"
	"os"
)

type Kademlia struct {
	network      *Network
	routingTable *RoutingTable
}

func initKademlia() {

	localAddr := os.Getenv("CONTAINER_IP")
	rootAddr := os.Getenv("ROOT_ADDRESS")

	// Create the network instance
	networkZero := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	me := NewContact(NewRandomKademliaID(), localAddr)
	rootContact := NewContact(NewRandomKademliaID(), rootAddr)

	routingTableZero := NewRoutingTable(me)

	// Initialize the Kademlia instance and set the network
	kademliaInstance := &Kademlia{
		network:      networkZero,
		routingTable: routingTableZero,
	}

	// Create a contact for the peer node
	if rootAddr != "" {
		networkZero.SendPingMessage(&rootContact)

	} else {
		fmt.Println("ROOT_ADDRESS not set in environment")
	}

}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
