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

	if localAddr == "" {
		// Handle the error (e.g., exit or provide a default IP)
		localAddr = "127.0.0.1" // Use a default value if environment variable is not set
	}

	// Start listening on the node's address
	//go kademlia.Listen("0.0.0.0", 8080)

	// Create a new Kademlia instance
	myKademlia := kademlia.NewKademlia(localAddr)

	// Check if the contact address is different from the local address
	if contactAddr != "" && contactAddr != localAddr {
		// Create a contact for the peer node
		id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
		contact := kademlia.NewContact(id, contactAddr)
		myKademlia.Network.SendPingMessage(&contact)
	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment or same as local")
	}

	// Keep the application running
	select {}
}
