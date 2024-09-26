package main

import (
	//"encoding/json"
	"fmt"
	"kademlia"
	//"net/http"
	"os"
)

func main() {
	fmt.Println("Starting Kademlia network...")

	// Retrieve the node's address and peer's contact address from environment variables
	localAddr := os.Getenv("CONTAINER_IP")
	contactAddr := os.Getenv("CONTACT_ADDRESS")
	if localAddr == "" {
		localAddr = "127.0.0.1"
	}

	// Create a new Kademlia instance
	myKademlia := kademlia.NewKademlia(localAddr)

	// Start listening on the node's address
	go myKademlia.Network.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if contactAddr != "" {
		contact := kademlia.NewContact(nil, contactAddr)
		myKademlia.Network.SendPing(&contact)
		myKademlia.LookupContact(myKademlia.Network.LocalID)
	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment")
	}


	select {}
}
