package kademlia

import (
	"fmt"
	"net"
	"strings"
)

type Network struct {
	LocalID    *KademliaID
	LocalAddr  string
	pingPongCh chan Msgs
}

func (network *Network) Listen(ip string, port int) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting the listener:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// parse in async parallell
		go network.parseConnection(conn)
	}
}

func (network *Network) parseConnection(conn net.Conn) {
	defer conn.Close()

	// Create a buffer to store the incoming data
	buffer := make([]byte, 1024) // could be 4096
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Convert buffer to string and trim whitespaces
	message := string(buffer[:n])
	message = strings.TrimSpace(message)

	switch {
	case strings.HasPrefix(message, "PING"):

		fmt.Println("Received NET message:", message)
		var pingOriginAddr, pingOriginID string
		_, err := fmt.Sscanf(message, "PING from %s %s", &pingOriginAddr, &pingOriginID)
		if err != nil {
			fmt.Println("Error parsing PING message:", err)
			return
		}
		network.pingPongCh <- Msgs{
			MsgsType: "PING",
			Contact: Contact{
				Address: pingOriginAddr,
				ID:      NewKademliaID(pingOriginID),
			},
		}
		// read from the channel
		pongMsgs := <-network.pingPongCh
		if pongMsgs.MsgsType == "PONG" {
			// Send the PONG response back to the original address
			network.sendPongResponse(pingOriginAddr)
		} else {
			fmt.Println("Unexpected message received:", pongMsgs)
		}

	case strings.HasPrefix(message, "PONG"):

		fmt.Println("Received NET message:", message)
		var pongOriginIP, pongOriginID string
		_, err := fmt.Sscanf(message, "PONG from %s %s", &pongOriginIP, &pongOriginID)
		if err != nil {
			fmt.Println("Error parsing PONG message:", err)
			return
		}
		network.pingPongCh <- Msgs{
			MsgsType: "PONG",
			Contact: Contact{
				Address: pongOriginIP,
				ID:      NewKademliaID(pongOriginID),
			},
		}

	default:
		fmt.Println("Received unknown message:", message)
	}
}

func (network *Network) SendPingMessage(remoteContact *Contact) {
	fmt.Println("Sending NET PING to remoteContact: ", remoteContact)
	conn, err := net.Dial("tcp", remoteContact.Address)
	if err != nil {
		fmt.Println("Error connecting to contact:", err)
		return
	}
	defer conn.Close()

	message := fmt.Sprintf("PING from %s %s", network.LocalAddr, network.LocalID.String())

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending ping message:", err)
	}
}

// sendPongResponse sends a PONG message back to the node that sent the PING
func (network *Network) sendPongResponse(pingOriginAddr string) {
	fmt.Println("Sending NET PONG to remoteContact: ", pingOriginAddr)
	// all nodes listen on port 8080, see main
	pongIP := net.JoinHostPort(pingOriginAddr, "8080")

	conn, err := net.Dial("tcp", pongIP)
	if err != nil {
		fmt.Println("Error connecting back to origin on port 8080:", err)
		return
	}
	defer conn.Close()

	// Send a PONG message back to the origin
	pongMessage := fmt.Sprintf("PONG from %s %s", network.LocalAddr, network.LocalID.String())
	_, err = conn.Write([]byte(pongMessage))
	if err != nil {
		fmt.Println("Error sending PONG message:", err)
		return
	}

	fmt.Println("Sent PONG message with server IP:", network.LocalAddr, "to", pongIP)
}

func handlePongMsgs(pongOriginIP string) {
	fmt.Println("inside handlePongMsgs with pongOriginIP: ", pongOriginIP)
}

// handleConnection handles incoming messages
func handleConnection(conn net.Conn) {
	fmt.Println("Handling connection from:", conn.RemoteAddr())
	defer conn.Close()
	// Handle incoming messages here TODO LATER
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
