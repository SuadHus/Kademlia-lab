package kademlia

import (
	"testing"
)

func TestRoutingTable_AddContact(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "localhost:8000")
	rt := NewRoutingTable(me)
	contact := NewContact(NewRandomKademliaID(), "localhost:8001")
	rt.AddContact(contact)

	// Verify that the contact is in the appropriate bucket
	bucketIndex := rt.getBucketIndex(contact.ID)
	bucket := rt.buckets[bucketIndex]
	if bucket.Len() == 0 {
		t.Error("Bucket should contain the contact")
	}
}

func TestRoutingTable_FindClosestContacts(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)

	contactsToAdd := []Contact{
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"),
		NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"),
		NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8003"),
		NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8004"),
		NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8005"),
		NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8006"),
	}

	for _, contact := range contactsToAdd {
		rt.AddContact(contact)
	}

	targetID := NewKademliaID("2111111400000000000000000000000000000000")
	contacts := rt.FindClosestContacts(targetID, 20)

	if len(contacts) != len(contactsToAdd) {
		t.Errorf("Expected %d contacts but got %d", len(contactsToAdd), len(contacts))
	}

	// Verify that contacts are sorted by distance
	for i := 1; i < len(contacts); i++ {
		if contacts[i-1].distance.Greater(contacts[i].distance) {
			t.Error("Contacts are not sorted by distance")
		}
	}
}

// You need to add Greater method to KademliaID
func (kademliaID *KademliaID) Greater(other *KademliaID) bool {
	for i := 0; i < IDLength; i++ {
		if kademliaID[i] != other[i] {
			return kademliaID[i] > other[i]
		}
	}
	return false
}
