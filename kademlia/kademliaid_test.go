package kademlia

import (
	"strings"
	"testing"
)

func TestNewKademliaID(t *testing.T) {
	idStr := "FFFFFFFF00000000000000000000000000000000"
	id := NewKademliaID(idStr)
	if !strings.EqualFold(id.String(), idStr) {
		t.Errorf("Expected ID string %s, got %s", idStr, id.String())
	}
}

func TestKademliaID_Less(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000001")
	id2 := NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	if !id1.Less(id2) {
		t.Error("Expected id1 to be less than id2")
	}
}

func TestKademliaID_Equals(t *testing.T) {
	id1 := NewKademliaID("1234567890ABCDEF1234567890ABCDEF12345678")
	id2 := NewKademliaID("1234567890ABCDEF1234567890ABCDEF12345678")
	if !id1.Equals(id2) {
		t.Error("Expected IDs to be equal")
	}
}

func TestKademliaID_CalcDistance(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	id2 := NewKademliaID("0000000000000000000000000000000000000000")
	distance := id1.CalcDistance(id2)
	expected := NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	if !distance.Equals(expected) {
		t.Error("Expected maximum distance")
	}
}
