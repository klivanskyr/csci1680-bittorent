package main

import (
	TrackingServer "bittorrent/pkg/trackingserver"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	tracker := TrackingServer.NewTracker()
	go tracker.Listen()

	// repl
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		input = strings.TrimSpace(input)
		switch input {
		case "help":
			fmt.Println("Commands:")
			fmt.Println("  help - display this message")
			fmt.Println("  lp - display the list of peers")
			fmt.Println("  exit - exit the program")
		case "lp":
			peersMap := tracker.GetPeers()
			fmt.Println("Peers:")
			for hash, peers := range peersMap {
				fmt.Println("  InfoHash:", hash)
				for _, peer := range peers {
					fmt.Print("    ", peer.IP, ":", peer.Port, "\n")
				}
			}
		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
		}
	}
}