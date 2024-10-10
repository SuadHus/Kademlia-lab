// kademlia_test.go
package kademlia

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// func TestKademlia_StoreAndRetrieveData(t *testing.T) {
// 	// Set up addresses and ports
// 	node1IP := "127.0.0.1"
// 	node1Port := 8000
// 	node1Addr := fmt.Sprintf("%s:%d", node1IP, node1Port)

// 	node2IP := "127.0.0.1"
// 	node2Port := 8001
// 	node2Addr := fmt.Sprintf("%s:%d", node2IP, node2Port)

// 	// Create two Kademlia nodes
// 	node1 := NewKademlia(node1Addr)
// 	node2 := NewKademlia(node2Addr)

// 	// Start listening on both nodes
// 	go node1.Network.Listen(node1IP, node1Port)
// 	go node2.Network.Listen(node2IP, node2Port)

// 	// Give some time for nodes to start listening
// 	time.Sleep(1 * time.Second)

// 	// Node1 joins the network via node2
// 	bootstrapContact := NewContact(node2.Network.LocalID, node2Addr)
// 	node1.JoinNetwork(&bootstrapContact)

// 	// Wait for the join to complete
// 	time.Sleep(1 * time.Second)

// 	// Store data on node1
// 	data := []byte("Hello Kademlia")
// 	err := node1.StoreData(data)
// 	if err != nil {
// 		t.Fatalf("Failed to store data: %v", err)
// 	}

// 	// Wait for data to be stored
// 	time.Sleep(500 * time.Millisecond)

// 	// Compute the key
// 	key := node1.HashData(data)

// 	// Retrieve data from node2
// 	retrievedData, err := node2.RetrieveData(key)
// 	if err != nil {
// 		t.Fatalf("Failed to retrieve data: %v", err)
// 	}

// 	// Verify the data matches
// 	if !bytes.Equal(data, retrievedData) {
// 		t.Fatalf("Retrieved data does not match stored data")
// 	}
// }

func TestKademlia_JoinNetwork(t *testing.T) {
	// Set up addresses and ports
	bootstrapIP := "127.0.0.1"
	bootstrapPort := 8001
	bootstrapAddr := fmt.Sprintf("%s:%d", bootstrapIP, bootstrapPort)

	nodeIP := "127.0.0.1"
	nodePort := 8000
	nodeAddr := fmt.Sprintf("%s:%d", nodeIP, nodePort)

	// Create bootstrap node
	bootstrapNode := NewKademlia(bootstrapAddr)
	go bootstrapNode.Network.Listen(bootstrapIP, bootstrapPort)

	// Create node that will join the network
	node := NewKademlia(nodeAddr)
	go node.Network.Listen(nodeIP, nodePort)

	// Give some time for nodes to start listening
	time.Sleep(1 * time.Second)

	// Node joins the network via bootstrap node
	bootstrapContact := NewContact(bootstrapNode.Network.LocalID, bootstrapAddr)
	node.JoinNetwork(&bootstrapContact)

	// Wait for join to complete
	time.Sleep(1 * time.Second)

	// Check if bootstrap node is in node's routing table
	responseCh := make(chan RoutingTableResponse)
	node.RoutingTableActionChannel <- RoutingTableAction{
		ActionType: "FindClosestContacts",
		TargetID:   bootstrapNode.Network.LocalID,
		ResponseCh: responseCh,
	}
	response := <-responseCh

	found := false
	for _, contact := range response.ClosestContacts {
		if contact.ID.Equals(bootstrapNode.Network.LocalID) {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Bootstrap node should be in the routing table")
	}
}

func TestKademlia_HandleMessages(t *testing.T) {
	// Set up addresses and ports
	nodeIP := "127.0.0.1"
	nodePort := 8000
	nodeAddr := fmt.Sprintf("%s:%d", nodeIP, nodePort)

	// Create a Kademlia node
	node := NewKademlia(nodeAddr)
	go node.Network.Listen(nodeIP, nodePort)

	// Give time for node to start listening
	time.Sleep(1 * time.Second)

	// Simulate receiving a PING message
	pingMessage := fmt.Sprintf("PING %s", node.Network.LocalID.String())
	response := node.HandleMessage(pingMessage, "127.0.0.1:8001")

	if !strings.HasPrefix(response, "PONG") {
		t.Fatalf("Expected PONG response, got: %s", response)
	}

	// Simulate receiving a FIND_NODE message
	targetID := NewRandomKademliaID()
	findNodeMessage := fmt.Sprintf("FIND_NODE %s", targetID.String())
	response = node.HandleMessage(findNodeMessage, "127.0.0.1:8001")

	if !strings.HasPrefix(response, "FIND_NODE_RESPONSE") {
		t.Fatalf("Expected FIND_NODE_RESPONSE, got: %s", response)
	}
}
