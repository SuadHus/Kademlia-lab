package main

import (
	"bufio"
	"fmt"
	"kademlia"
	"os"
	"strings"
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

	// Join the network if a contact address is provided
	if contactAddr != "" {
		id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
		contact := kademlia.NewContact(id, contactAddr)
		myKademlia.JoinNetwork(&contact)
		myKademlia.PrintRoutingTable()
	} else {
		fmt.Println("CONTACT_ADDRESS not set in environment")
	}

	// keep the node running
	select {}
}

func cli(myKademlia *kademlia.Kademlia) {

	// Start CLI interaction
	fmt.Println("Kademlia CLI started. Available commands: put <data>, get <hash>, join <address>, quit")

	// Read user input in an infinite loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {

			continue
		}
		input = strings.TrimSpace(input)
		if input == "quit" {
			fmt.Println("Exiting...")
			break
		}

		// Handle CLI commands
		handleCommand(input, myKademlia)
	}
}

func handleCommand(input string, myKademlia *kademlia.Kademlia) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	switch command {
	case "put":
		if len(parts) < 2 {
			fmt.Println("Usage: put <data>")
			return
		}
		data := strings.Join(parts[1:], " ")
		err := myKademlia.StoreData([]byte(data))
		if err != nil {
			fmt.Println("Failed to store data:", err)
			return
		}
		hash := myKademlia.HashData([]byte(data))
		fmt.Println("Data stored with hash:", hash)

	case "get":
		if len(parts) < 2 {
			fmt.Println("Usage: get <hash>")
			return
		}
		hash := parts[1]
		data, err := myKademlia.RetrieveData(hash)
		if err != nil {
			fmt.Println("Failed to retrieve data:", err)
			return
		}
		fmt.Println("Data retrieved:", string(data))

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Available commands: put <data>, get <hash>, join <address>, quit")
	}
}
