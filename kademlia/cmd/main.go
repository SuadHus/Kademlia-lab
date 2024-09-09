package main

import (
	"fmt"
	"kademlia"
	"os"
)

func main() {
	fmt.Println("Starting Kademlia network...")

	// Retrieve the node's address and peer's contact address from environment variables
	localAddr := os.Getenv("NODE_ADDRESS")
	contactAddr := os.Getenv("CONTACT_ADDRESS")

	if localAddr == "" || contactAddr == "" {
		fmt.Println("NODE_ADDRESS or CONTACT_ADDRESS not set in environment")
		return
	}

	// Example network initialization for this node
	network := kademlia.Network{
		LocalID:   kademlia.NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	// Start listening on the node's address
	go kademlia.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), contactAddr)

	// Send a ping to the peer node
	network.SendPingMessage(&contact)

	// Keep the application running
	select {}
}
