package main

import (
	//"encoding/json"
	"fmt"
	"kademlia"
	//"net/http"
	"os"
	"time"
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
		id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
		contact := kademlia.NewContact(id, contactAddr)
		
		myKademlia.Network.SendPing(&contact)
		time.Sleep(2 * time.Second)
		myKademlia.LookupContact(myKademlia.Network.LocalID)

		time.Sleep(2 * time.Second)
		myKademlia.PrintRoutingTable()
	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment")
	}


	select {}
}
