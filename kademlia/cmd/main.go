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

	localAddr := os.Getenv("CONTAINER_IP")
	contactAddr := os.Getenv("CONTACT_ADDRESS")
	if localAddr == "" {
		localAddr = "127.0.0.1"
	}

	myKademlia := kademlia.NewKademlia(localAddr)

	go myKademlia.Network.Listen("0.0.0.0", 8080)

	// Create a contact for the peer node
	if contactAddr != "" {
		id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000")
		contact := kademlia.NewContact(id, contactAddr)

		myKademlia.JoinNetwork(&contact)

		myKademlia.PrintRoutingTable()
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
		hash := myKademlia.StoreData([]byte(data))
		fmt.Fprintf(w, "Data stored with hash: %s", hash)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hash := r.URL.Query().Get("hash")
		data, err := myKademlia.RetrieveData(hash)
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
