package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Starting client")
	tmuxServerSocketPath := "./nc.sock"

	// Connect to the tmux server
	conn, err := net.Dial("unix", tmuxServerSocketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("connected to", tmuxServerSocketPath)

	// Read response
	for {
		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read: %v\n", err)
			return
		}

		fmt.Println("server replied:", response)
	}
}
