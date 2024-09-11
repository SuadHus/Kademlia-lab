package main

import (
	"fmt"
	"kademlia"
	"os"
)

func main() {
	//fmt.Println("Starting Kademlia network...")

	// Retrieve the node's address and peer's contact address from environment variables
	localAddr := os.Getenv("CONTAINER_IP")
	rootAddr := os.Getenv("ROOT_ADDRESS")

	if localAddr == "" || rootAddr == "" {
		fmt.Println("NODE_ADDRESS or ROOT_ADDRESS not set in environment")
	}

	// Example network initialization for this node
	network := kademlia.Network{
		LocalID:   kademlia.NewRandomKademliaID(),
		LocalAddr: localAddr,
	}

	// Start listening on the node's address
	go kademlia.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if rootAddr != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), rootAddr)
		network.SendPingMessage(&contact)

	} else {
		fmt.Println("ROOT_ADDRESS not set in environment")
	}

	// Send a ping to the root node
	//contact := kademlia.Contact{Address: rootAddr}
	//network.SendPingMessage(&contact)

	// Keep the application running
	select {}
}
