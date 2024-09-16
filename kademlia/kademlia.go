package kademlia

import (
	"fmt"
)

type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
	Me           *Contact
	pingPongCh   chan Msgs
}

type Msgs struct {
	MsgsType string
	Contact  Contact
}

func InitKademlia(localAddr string) *Kademlia {

	pingPongCh := make(chan Msgs)

	networkZero := &Network{
		LocalID:    NewRandomKademliaID(),
		LocalAddr:  localAddr,
		pingPongCh: pingPongCh,
	}

	me := NewContact(NewRandomKademliaID(), localAddr)
	routingTableZero := NewRoutingTable(me)
	fmt.Println("Kademliainstance Created with routingtable:", routingTableZero)

	// Initi kademlia and set its network
	return &Kademlia{
		Network:      networkZero,
		RoutingTable: routingTableZero,
		Me:           &me,
		pingPongCh:   pingPongCh,
	}

}

func (k *Kademlia) ListenForMessages() {
	go func() {
		for msgs := range k.pingPongCh {
			switch msgs.MsgsType {
			case "PING":
				fmt.Println("Received PING from contact:", msgs.Contact)
				pongMsg := Msgs{
					MsgsType: "PONG",
					Contact:  *k.Me, // send me back in the PONG
				}
				k.pingPongCh <- pongMsg
			case "PONG":
				fmt.Println("Received PONG from contact:", msgs.Contact)
				k.RoutingTable.AddContact(msgs.Contact)
			default:
				fmt.Println("Received unknown message type:", msgs.MsgsType)
			}
		}
	}()
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
