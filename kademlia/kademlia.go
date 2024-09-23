package kademlia

import (
	"fmt"
	"sort"
)

const alpha = 3

type Kademlia struct {
	Network      *Network
	RoutingTable *RoutingTable
	cmdChannel   chan CmdChMsgs  // channel to pass commands from network module to kademlia
	dataChannel  chan DataChMsgs // channel to pass async data back from kademlia to network module
}

type CmdChMsgs struct {
	KademCmd   string // Command for kademlia to execute
	SenderID   string // sender kademlia ID
	SenderAddr string // sender IP
	TargetID   string // target ID for FIND_NODE
}

type DataChMsgs struct {
	NetCmd   string    // command for network, not sure if needed yet
	Contacts []Contact // contact data
	Data     []byte    // file data

}

func InitKademlia(localAddr string, rootAddr string) *Kademlia {

	cmdChannel := make(chan CmdChMsgs)
	dataChannel := make(chan DataChMsgs)

	var me Contact
	if localAddr == "172.16.238.10" {
		fmt.Println("Local address: ", localAddr)
		me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), localAddr)
	} else {
		me = NewContact(NewRandomKademliaID(), localAddr)
	}

	myNetwork := &Network{
		LocalID:     me.ID,
		LocalAddr:   localAddr,
		cmdChannel:  cmdChannel,
		dataChannel: dataChannel,
	}

	myRoutingTable := NewRoutingTable(me)

	k := &Kademlia{
		Network:      myNetwork,
		RoutingTable: myRoutingTable,
		cmdChannel:   cmdChannel,
		dataChannel:  dataChannel,
	}

	go k.Network.Listen("0.0.0.0", 8080)
	go k.ListenForChMsgs()

	return k
}

// LookupContact performs an iterative node lookup
func (kademlia *Kademlia) LookupContact(targetID *KademliaID) []Contact {
	alpha := 3
	k := bucketSize

	// Initialize the shortlist with k closest contacts
	closestContacts := kademlia.RoutingTable.FindClosestContacts(targetID, k) // channel fo reading the RT
	shortlist := make(map[string]Contact)
	queried := make(map[string]bool)

	for _, contact := range closestContacts {
		shortlist[contact.ID.String()] = contact
	}

	for {
		// Find the α closest unqueried contacts
		var unqueriedContacts []Contact
		for _, contact := range shortlist {
			if !queried[contact.ID.String()] {
				unqueriedContacts = append(unqueriedContacts, contact)
			}
		}

		if len(unqueriedContacts) == 0 {
			break
		}

		// Sort unqueriedContacts by distance to target
		sort.Slice(unqueriedContacts, func(i, j int) bool {
			distI := unqueriedContacts[i].ID.CalcDistance(targetID)
			distJ := unqueriedContacts[j].ID.CalcDistance(targetID)
			return distI.Less(distJ)
		})

		// Select up to α contacts to query
		numToQuery := alpha
		if len(unqueriedContacts) < alpha {
			numToQuery = len(unqueriedContacts)
		}
		contactsToQuery := unqueriedContacts[:numToQuery]

		// For each contact to query
		for _, contact := range contactsToQuery {
			queried[contact.ID.String()] = true

			// Send FIND_NODE to the contact
			contactsReceived := kademlia.SendFindNode(targetID, &contact) //channel operation

			// Add the contact to the routing table
			kademlia.RoutingTable.AddContact(contact)

			// Add any new contacts received to the shortlist
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

	// Convert shortlist to a slice and sort
	var finalContacts []Contact
	for _, contact := range shortlist {
		finalContacts = append(finalContacts, contact)
	}

	// Sort the finalContacts by distance to target
	sort.Slice(finalContacts, func(i, j int) bool {
		distI := finalContacts[i].ID.CalcDistance(targetID)
		distJ := finalContacts[j].ID.CalcDistance(targetID)
		return distI.Less(distJ)
	})

	// Return the k closest contacts
	if len(finalContacts) > k {
		finalContacts = finalContacts[:k]
	}

	kademlia.PrintAllContacts(kademlia.RoutingTable)

	return finalContacts
}

// SendFindNode sends a FIND_NODE message to a contact
func (kademlia *Kademlia) SendFindNode(targetID *KademliaID, contact *Contact) []Contact {
	contacts, err := kademlia.Network.SendFindNode(contact, targetID)
	if err != nil {
		fmt.Println("Error during SendFindNode:", err)
		return nil
	}
	return contacts
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}

func (k *Kademlia) ListenForChMsgs() {
	go func() {
		for msgs := range k.cmdChannel {
			switch msgs.KademCmd {

			case "PING":
				fmt.Println("Network module sent CH msgs about PING from:", msgs.SenderAddr)
				contact := NewContact(NewKademliaID(msgs.SenderID), msgs.SenderAddr)
				k.RoutingTable.AddContact(contact)
				go k.PrintAllContacts(k.RoutingTable)

			case "PONG":
				fmt.Println("Received PONG from:", msgs.SenderAddr)
				contact := NewContact(NewKademliaID(msgs.SenderID), msgs.SenderAddr)
				k.RoutingTable.AddContact(contact)
				go k.PrintAllContacts(k.RoutingTable)

			case "FIND_NODE":
				fmt.Println("Network module sent CH msgs about FIND_NODE from:", msgs.SenderAddr)
				targetID := NewKademliaID(msgs.TargetID)

				// Perform the lookup
				closestContacts := k.RoutingTable.FindClosestContacts(targetID, bucketSize)

				fmt.Println("Closest contacts to target ID:", closestContacts)

				dataMsgs := DataChMsgs{
					Contacts: closestContacts,
				}
				k.dataChannel <- dataMsgs

				go k.PrintAllContacts(k.RoutingTable)

			default:
				fmt.Println("Received unknown message type:", msgs.KademCmd)
			}
		}
	}()
}

func (k *Kademlia) PrintAllContacts(rt *RoutingTable) {
	fmt.Println("Contacts in the Routing Table:")
	for i, bucket := range rt.buckets {
		if bucket == nil || bucket.Len() == 0 {
			continue
		}
		fmt.Printf("Bucket %d:\n", i)
		for e := bucket.list.Front(); e != nil; e = e.Next() {
			contact := e.Value.(Contact)
			fmt.Printf("Contact ID: %s, Address: %s\n", contact.ID.String(), contact.Address)
			//fmt.Printf("Contact ID: %s, Address: %s, Distance: %s\n", contact.ID.String(), contact.Address, contact.distance.String()) //also print distance
		}
	}
}
