package main

import (
	"fmt"
	"kademlia"
	"os"
	"time"
)

func main() {

	localAddr := os.Getenv("CONTAINER_IP")
	rootAddr := os.Getenv("ROOT_ADDRESS")

	myKademlia := kademlia.InitKademlia(localAddr, rootAddr) // init kademlia

	if rootAddr != "" && rootAddr != localAddr {
		id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
		contact := kademlia.NewContact(id, rootAddr)
		myKademlia.Network.SendPingMessage(&contact)
		time.Sleep(1 * time.Second)

		fmt.Println("hallå", myKademlia.LookupContact(myKademlia.RoutingTable.Me.ID))
		//fmt.Println(myKademlia.Network.SendFindNode(&contact, id))

	} else {
		fmt.Println("root node does not init ping")
	}

	// Keep the application running
	select {}
}
