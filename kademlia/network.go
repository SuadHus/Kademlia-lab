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
	remoteIP := conn.RemoteAddr().String()

	// Create a buffer to store the incoming data
	buffer := make([]byte, 1024)
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
		fmt.Println("Received PING message:", message, "FROM: ", remoteIP)
		network.pingPongCh <- Msgs{
			MsgsType: "PING",
			Contact: Contact{
				Address: remoteIP,
			},
		}
		fmt.Println("before SendPongResponse Call")
		network.sendPongResponse(remoteIP)

	case strings.HasPrefix(message, "PONG"):
		fmt.Println("Received PONG message:", message)
		network.pingPongCh <- Msgs{
			MsgsType: "PONG",
			Contact: Contact{
				Address: remoteIP,
			},
		}

	default:
		fmt.Println("Received unknown message:", message)
	}
}

// sendPongResponse sends a PONG message back to the node that sent the PING
func (network *Network) sendPongResponse(pingOriginAddr string) {

	pongIP, _, err := net.SplitHostPort(pingOriginAddr)
	if err != nil {
		fmt.Println("Error extracting IP from origin address:", err)
		return
	}

	// all nodes listen on port 8080, see main
	pongIP = net.JoinHostPort(pongIP, "8080")

	conn, err := net.Dial("tcp", pongIP)
	if err != nil {
		fmt.Println("Error connecting back to origin on port 8080:", err)
		return
	}
	defer conn.Close()

	// Send a PONG message back to the origin
	pongMessage := "PONG from: " + network.LocalAddr
	_, err = conn.Write([]byte(pongMessage))
	if err != nil {
		fmt.Println("Error sending PONG message:", err)
		return
	}

	fmt.Println("Sent PONG message with server IP:", network.LocalAddr, "to", pongIP)
}

// SendPingMessage sends a ping message to the given contact
func (network *Network) SendPingMessage(contact *Contact) {
	conn, err := net.Dial("tcp", contact.Address)
	if err != nil {
		fmt.Println("Error connecting to contact:", err)
		return
	}
	defer conn.Close()

	message := fmt.Sprintf("PING from %s", network.LocalID.String())
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending ping message:", err)
	}
}

// func (network *Network) SendBootstrapInitPing(target Contact) {
// 	botstrapPing := Msgs{
// 		MsgsType: "PING",
// 		Contact:  target,
// 	}
// 	network.pingPongCh <- botstrapPing
// }

// func handlePingMsgs(pingOriginAddr string) {
// 	pingOriginIP, _, err := net.SplitHostPort(pingOriginAddr)
// 	if err != nil {
// 		fmt.Println("Error extracting IP from origin address:", err)
// 		return
// 	}

// 	// Set the PONG response to be sent to the PING origin IP on port 8080
// 	pingOriginWithPort := net.JoinHostPort(pingOriginIP, "8080")

// 	conn, err := net.Dial("tcp", pingOriginWithPort)
// 	if err != nil {
// 		fmt.Println("Error connecting back to origin IP on port 8080:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	localAddr := conn.LocalAddr().String()
// 	pongMessage := "PONG from: " + localAddr
// 	_, err = conn.Write([]byte(pongMessage))
// 	if err != nil {
// 		fmt.Println("Error sending PONG response:", err)
// 		return
// 	}

// 	fmt.Println("Sent PONG message with server IP:", localAddr, "to", pingOriginWithPort)
// }

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
