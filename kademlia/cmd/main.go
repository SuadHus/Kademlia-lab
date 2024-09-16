package main

import (
	"encoding/json"
	"fmt"
	"kademlia"
	"net/http"
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

	// Set up HTTP handlers
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data := r.FormValue("data")
		hash := myKademlia.Store([]byte(data))
		fmt.Fprintf(w, "Data stored with hash: %s", hash)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hash := r.URL.Query().Get("hash")
		data, err := myKademlia.LookupData(hash)
		if err != nil {
			http.Error(w, "Data not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"data": string(data)})
	})

	// Start HTTP server
	fmt.Println("HTTP server listening on port 8001")
	http.ListenAndServe(":8001", nil)

	select {}
}
