package kademlia

import "fmt"

// Kademlia represents the Kademlia distributed hash table.
type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
}

// NewKademlia initializes a new Kademlia instance.
func NewKademlia(localAddr string) *Kademlia {
	network := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	newContact := NewContact(NewRandomKademliaID(), localAddr)

	routingTable := NewRoutingTable(newContact)
	fmt.Println("New Kademlia instance created", routingTable)
	return &Kademlia{
		Network:      network,
		RoutingTable: routingTable,
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
