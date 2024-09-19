package kademlia

import (
	"fmt"
)

type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
	pingPongCh   chan ChMsgs
}

type ChMsgs struct {
	MsgsType   string
	SenderID   string
	SenderAddr string
}

func InitKademlia(localAddr string, rootAddr string) *Kademlia {
	pingPongCh := make(chan ChMsgs)

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

	k := &Kademlia{
		Network:      myNetwork,
		RoutingTable: myRoutingTable,
		pingPongCh:   pingPongCh,
	}

	go k.Network.Listen("0.0.0.0", 8080)
	go k.ListenForMsgs()

	return k
}

func (k *Kademlia) ListenForMsgs() {
	go func() {
		for msgs := range k.pingPongCh {
			switch msgs.MsgsType {

			case "PING":
				fmt.Println("Network module sent CH msgs about PING from: ", msgs.SenderAddr)
				contact := NewContact(NewKademliaID(msgs.SenderID), msgs.SenderAddr)
				k.RoutingTable.AddContact(contact)
				go PrintAllContacts(k.RoutingTable)

				// Send PONG response
				pongMsg := ChMsgs{
					MsgsType:   "PONG",
					SenderID:   k.RoutingTable.me.ID.String(),
					SenderAddr: k.RoutingTable.me.Address,
				}
				k.pingPongCh <- pongMsg

			case "PONG":
				fmt.Println("Network module sent CH msgs about PONG from: ", msgs.SenderAddr)
				contact := NewContact(NewKademliaID(msgs.SenderID), msgs.SenderAddr)
				k.RoutingTable.AddContact(contact)
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
