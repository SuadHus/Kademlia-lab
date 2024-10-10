package kademlia

import (
	"testing"
)

func TestNewContact(t *testing.T) {
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	address := "localhost:8000"
	contact := NewContact(id, address)
	if !contact.ID.Equals(id) {
		t.Error("Contact ID does not match")
	}
	if contact.Address != address {
		t.Error("Contact address does not match")
	}
}

func TestContact_CalcDistance(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("00000000FFFFFFFF000000000000000000000000")
	contact := NewContact(id1, "localhost:8000")
	contact.CalcDistance(id2)
	expected := id1.CalcDistance(id2)
	if !contact.distance.Equals(expected) {
		t.Error("Contact distance calculation is incorrect")
	}
}

func TestContactCandidates_Sort(t *testing.T) {
	targetID := NewKademliaID("ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCD")
	contact1 := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), "addr1")
	contact2 := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "addr2")
	contact3 := NewContact(targetID, "addr3")
	contacts := []Contact{contact1, contact2, contact3}

	candidates := ContactCandidates{}
	candidates.Append(contacts)

	// Calculate distances
	for i := range candidates.contacts {
		candidates.contacts[i].CalcDistance(targetID)
	}

	candidates.Sort()

	if !candidates.contacts[0].ID.Equals(targetID) {
		t.Error("Expected contact3 to be closest to target")
	}
}
