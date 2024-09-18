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

	var id *kademlia.KademliaID
	if localAddr == rootAddr {
		id = kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	} else {
		id = kademlia.NewRandomKademliaID()
	}

	myKademlia := kademlia.InitKademlia(localAddr, *id) // init kademlia
	myKademlia.ListenForMsgs()                          // set kademlia to listen for msgs on channel
	go myKademlia.Network.Listen("0.0.0.0", 8080)       // every node listens on incomeing net traffic on 8080

	// Create a contact for the peer node
	if rootAddr != "" {
		rootContact := kademlia.NewContact(id, rootAddr)
		myKademlia.Network.SendPingMessage2(&rootContact)
	} else {
		fmt.Println("Root node does not init ping")
	}

	fmt.Println(myKademlia.RoutingTable)

	// Keep the application running
	select {}
}
