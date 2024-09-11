package main

import (
	"fmt"
	"kademlia"
	"os"
)

func main() {
	fmt.Println("Starting Kademlia network...")

	// Retrieve the node's address and peer's contact address from environment variables
	localAddr := os.Getenv("CONTAINER_IP")
	contactAddr := os.Getenv("CONTACT_ADDRESS")
	root_node := os.Getenv("ROOT_NODE")

	if localAddr == "" || contactAddr == "" {
		fmt.Println("NODE_ADDRESS or CONTACT_ADDRESS not set in environment")
	}

	// Example network initialization for this node
	network := kademlia.Network{
		LocalID:   kademlia.NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	// Start listening on the node's address
	go kademlia.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if contactAddr != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), contactAddr)
		if root_node != "" {
			network.SendPingMessage(&contact)
		}

	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment")
	}

	// Send a ping to the peer node

	// Keep the application running
	select {}
}
