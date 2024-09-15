package kademlia

import (
	"fmt"
)

type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
}

func InitKademlia(localAddr string) *Kademlia {

	networkZero := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	me := NewContact(NewRandomKademliaID(), localAddr)
	routingTableZero := NewRoutingTable(me)
	fmt.Println("Kademliainstance Created with routingtable:", routingTableZero)

	// Initi kademlia and set its network
	return &Kademlia{
		Network:      networkZero,
		RoutingTable: routingTableZero,
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
