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
	//contactAddr := os.Getenv("CONTACT_ADDRESS")
	if localAddr == "" {
		// handle the error (e.g., exit or provide a default IP)
		localAddr = "127.0.0.1" // Use a default value if environment variable is not set
	}

	// Create a new Kademlia instance
	myKademlia := kademlia.NewKademlia(localAddr)

	// Start listening on the node's address
	go kademlia.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if contactAddr != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), contactAddr)
		myKademlia.Network.SendPingMessage(&contact)

	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment")
	}

	// Send a ping to the peer node

	// Keep the application running
	select {}
}
