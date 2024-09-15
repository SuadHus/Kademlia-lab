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
		localAddr = "127.0.0.1"
	}

	// Create a new Kademlia instance
	myKademlia := kademlia.NewKademlia(localAddr)

	// Start listening on the node's address
	go myKademlia.Network.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if contactAddr != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), contactAddr)
		myKademlia.SendPing(&contact)
	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment")
	}

	// Command line interface
	// reader := bufio.NewReader(os.Stdin)
	// for {
	// 	fmt.Print("Enter command (put/get/exit): ")
	// 	text, _ := reader.ReadString('\n')
	// 	text = strings.TrimSpace(text)
	// 	if text == "exit" {
	// 		fmt.Println("Exiting...")
	// 		os.Exit(0)
	// 	} else if strings.HasPrefix(text, "put ") {
	// 		data := strings.TrimPrefix(text, "put ")
	// 		dataBytes := []byte(data)
	// 		hash := myKademlia.Store(dataBytes)
	// 		fmt.Println("Data stored with hash:", hash)
	// 	} else if strings.HasPrefix(text, "get ") {
	// 		hash := strings.TrimPrefix(text, "get ")
	// 		data, err := myKademlia.LookupData(hash)
	// 		if err != nil {
	// 			fmt.Println("Error retrieving data:", err)
	// 		} else {
	// 			fmt.Println("Data retrieved:", string(data))
	// 		}
	// 	} else {
	// 		fmt.Println("Unknown command:", text)
	// 	}
	// }

	select {}

}
