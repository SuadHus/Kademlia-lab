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

func InitKademlia(localAddr string, rootAddr string) *Kademlia {
	pingPongCh := make(chan Msgs)

	var me Contact
	if localAddr == "172.16.238.10" {
		fmt.Println("Local address: ", localAddr)
		me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), localAddr)
	} else {
		me = NewContact(NewRandomKademliaID(), localAddr)
	}

	myNetwork := &Network{
		LocalID:    me.ID,
		LocalAddr:  localAddr,
		pingPongCh: pingPongCh,
	}

	myRoutingTable := NewRoutingTable(me)

	// Initi kademlia and set its network
	k := &Kademlia{
		Network:      myNetwork,
		RoutingTable: myRoutingTable,
		pingPongCh:   pingPongCh,
	}

	go k.Network.Listen("0.0.0.0", 8080) // every node listens on incomeing net traffic on 8080
	go k.ListenForMsgs()                 // set kademlia to listen for msgs on channel

	return k
}

func (k *Kademlia) ListenForMsgs() {
	go func() {
		for msgs := range k.pingPongCh {
			switch msgs.MsgsType {

			case "PING":
				fmt.Println("Network module sent CH msgs about PING from contact:", msgs.Contact)
				k.RoutingTable.AddContact(msgs.Contact) // add contact of incoming ping to rt
				go PrintAllContacts(k.RoutingTable)
				pongMsg := Msgs{
					MsgsType: "PONG",
					Contact:  k.RoutingTable.me, // send me back in the PONG

				}
				k.pingPongCh <- pongMsg

			case "PONG":
				fmt.Println("Network module sent CH msgs about PONG from contact:", msgs.Contact)
				k.RoutingTable.AddContact(msgs.Contact)
				go PrintAllContacts(k.RoutingTable)

			default:
				fmt.Println("Received unknown message type:", msgs.MsgsType)
			}
		}
	}()
}

func PrintAllContacts(rt *RoutingTable) {
	fmt.Println("Contacts in the Routing Table:")
	for i, bucket := range rt.buckets {
		if bucket == nil || bucket.Len() == 0 {
			continue
		}
		fmt.Printf("Bucket %d:\n", i)
		for e := bucket.list.Front(); e != nil; e = e.Next() {
			contact := e.Value.(Contact)
			fmt.Printf("Contact ID: %s, Address: %s\n", contact.ID.String(), contact.Address)
		}
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
