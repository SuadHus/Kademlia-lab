package kademlia

import (
	"fmt"
	"strings"
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
			// Split the message by spaces to parse it
			parts := strings.Split(msg, " ")
			if len(parts) == 4 && (parts[0] == "PING" || parts[0] == "PONG") && parts[1] == "from" {
				// Extract the Kademlia ID and address from the message
				contactID := NewKademliaID(parts[2]) // The 3rd part is the ID
				contactAddr := parts[3]              // The 4th part is the Address

				// Create a new contact and add it to the routing table
				newContact := NewContact(contactID, contactAddr)
				kademlia.RoutingTable.AddContact(newContact)

				fmt.Println("Received message on Kademlia channel:", msg)
				go PrintAllContacts(kademlia.RoutingTable)

			} else {
				fmt.Println("Invalid message format:", msg)
			}

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
