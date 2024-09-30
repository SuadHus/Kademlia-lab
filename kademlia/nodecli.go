package kademlia

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type NodeCli struct {
	nodeCli CliMessageHandler
}

// MessageHandler interface to pass cli commands to kademlia
type CliMessageHandler interface {
	HandleCliMessage(message string, senderAddr string) string
}

func (cli *NodeCli) StartCLI() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Kademlia CLI started. Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		trimmedInput := strings.TrimSpace(input)

		if trimmedInput == "exit" {
			fmt.Println("Exiting Kademlia Node CLI...")
			break
		}

		if trimmedInput == "help" {
			fmt.Println("Commands:\n - put <data> : Store data in the Kademlia network\n - get <hash> : Retrieve data using its hash")
			continue
		}

		if strings.HasPrefix(trimmedInput, "put ") {
			dataToStore := strings.TrimPrefix(trimmedInput, "put ")
			response := cli.nodeCli.HandleCliMessage("put "+dataToStore, "")
			fmt.Println(response)
			continue
		}

		if strings.HasPrefix(trimmedInput, "get ") {
			hashToRetrieve := strings.TrimPrefix(trimmedInput, "get ")
			response := cli.nodeCli.HandleCliMessage("get "+hashToRetrieve, "")
			fmt.Println(response)
			continue
		}

		// default
		fmt.Println("Unknown command. Type 'help' for a list of commands.")
	}
}
