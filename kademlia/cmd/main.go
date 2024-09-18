package main

import (
	"fmt"
	"kademlia"
	"os"
)

func main() {

	localAddr := os.Getenv("CONTAINER_IP")
	rootAddr := os.Getenv("ROOT_ADDRESS")

	myKademlia := kademlia.InitKademlia(localAddr, rootAddr) // init kademlia

	if rootAddr != "" && rootAddr != localAddr {
		id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
		contact := kademlia.NewContact(id, rootAddr)
		myKademlia.Network.SendPingMessage(&contact)
	} else {
		fmt.Println("ROOT_ADDRESS not set in environment or same as local")
	}

	// Keep the application running
	select {}
}
