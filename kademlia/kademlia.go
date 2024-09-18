package kademlia

import (
	"fmt"
)

type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
	pingPongCh   chan Msgs
}

type Msgs struct {
	MsgsType string
	Contact  Contact
}

func InitKademlia(localAddr string, id KademliaID) *Kademlia {
	pingPongCh := make(chan Msgs)

	myNetwork := &Network{
		LocalID:    &id,
		LocalAddr:  localAddr,
		pingPongCh: pingPongCh,
	}

	me := NewContact(&id, localAddr)
	myRoutingTable := NewRoutingTable(me)

	// Initi kademlia and set its network
	return &Kademlia{
		Network:      myNetwork,
		RoutingTable: myRoutingTable,
		pingPongCh:   pingPongCh,
	}

}

func (k *Kademlia) ListenForMsgs() {
	go func() {
		for msgs := range k.pingPongCh {
			switch msgs.MsgsType {
			case "PING":
				fmt.Println("Received PING from contact:", msgs.Contact, "sending Pong to kadem channel")
				pongMsg := Msgs{
					MsgsType: "PONG",
					Contact:  k.RoutingTable.me, // send me back in the PONG
				}
				k.pingPongCh <- pongMsg
			case "PONG":
				fmt.Println("Received PONG from contact:", msgs.Contact)
				k.RoutingTable.AddContact(msgs.Contact)
				fmt.Println("added contact: ", msgs.Contact, "too routing table")
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
