package kademlia

import (
	"strings"
	"testing"
	"time"
)

type MockHandler struct{}

func (mh *MockHandler) HandleMessage(message string, senderAddr string) string {
	if strings.HasPrefix(message, "PING ") {
		// Extract the sender ID (though it's not necessary here)
		// Respond with "PONG <some ID>"
		return "PONG MOCK_ID"
	}
	return ""
}

func TestNetwork_SendPing(t *testing.T) {
	localID := NewRandomKademliaID()
	localAddr := "127.0.0.1:9000"
	network := &Network{
		LocalID:   localID,
		LocalAddr: localAddr,
		handler:   &MockHandler{},
	}

	// Start a mock listener
	go network.Listen("127.0.0.1", 9000)
	time.Sleep(100 * time.Millisecond) // Give the listener time to start

	contact := NewContact(localID, localAddr)
	err := network.SendPing(&contact)
	if err != nil {
		t.Errorf("Failed to send PING: %v", err)
	}
}
