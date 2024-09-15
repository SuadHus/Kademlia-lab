package main

import (
	"fmt"
	"kademlia"
	"os"
)

func main() {

	localAddr := os.Getenv("CONTAINER_IP")
	rootAddr := os.Getenv("ROOT_ADDRESS")

	if localAddr == "" {
		localAddr = "127.0.0.1" // Use a default value if environment variable is not set
	}

	myKademlia := kademlia.InitKademlia(localAddr)

	// Start listening for incoming msgs on port 8080
	go kademlia.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if rootAddr != "" {
		rootContact := kademlia.NewContact(kademlia.NewRandomKademliaID(), rootAddr)
		myKademlia.Network.SendPingMessage(&rootContact)
	} else {
		fmt.Println("Root node does not init ping")
	}

	// Keep the application running
	select {}
}
