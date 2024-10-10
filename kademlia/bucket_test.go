package kademlia

import (
	"testing"
)

func TestBucket_AddContact(t *testing.T) {
	bucket := newBucket()
	contact1 := NewContact(NewRandomKademliaID(), "localhost:8000")
	contact2 := NewContact(NewRandomKademliaID(), "localhost:8001")

	bucket.AddContact(contact1)
	bucket.AddContact(contact2)

	if bucket.Len() != 2 {
		t.Errorf("Expected bucket length 2, got %d", bucket.Len())
	}

	// Test adding the same contact again
	bucket.AddContact(contact1)
	if bucket.Len() != 2 {
		t.Errorf("Bucket should not add duplicate contacts, got length %d", bucket.Len())
	}
}

func TestBucket_GetContactAndCalcDistance(t *testing.T) {
	bucket := newBucket()
	targetID := NewRandomKademliaID()
	contact := NewContact(NewRandomKademliaID(), "localhost:8000")
	bucket.AddContact(contact)
	contacts := bucket.GetContactAndCalcDistance(targetID)
	if len(contacts) != 1 {
		t.Error("Expected 1 contact")
	}
	if contacts[0].distance == nil {
		t.Error("Contact distance should be calculated")
	}
}
