package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func main() {
	// Define the handler function for the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get the container hostname
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		// Get the container IP address
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			http.Error(w, "Unable to get IP", http.StatusInternalServerError)
			return
		}
		var ipAddr string
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipAddr = ipNet.IP.String()
					break
				}
			}
		}

		// Respond with the hostname and IP address
		fmt.Fprintf(w, "Hello from container %s! My IP address is %s.\n", hostname, ipAddr)
	})

	// Start the HTTP server on port 8080
	fmt.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
	}
}
