package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	tmuxClientSocketPath := "./guard.sock"

	//Clean up old socket
	_, err := os.Stat(tmuxClientSocketPath)
	if err == nil {
		err = os.Remove(tmuxClientSocketPath)
		if err != nil {
			log.Fatal("Tried to clear out old socket but failed")
		}
	}

	// Create a Unix domain socket and listen for incoming connections.
	clientSocket, err := net.Listen("unix", tmuxClientSocketPath)
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(tmuxClientSocketPath)
		os.Exit(1)
	}()

	for {
		// Accept an incoming connection.
		conn, err := clientSocket.Accept()
		if err != nil {
			fmt.Println("Client had a problem connecting")
			fmt.Println(err)
		}

		// Handle the connection in a separate goroutine.
		go run_client(conn)
	}
}

func run_client(clientCon net.Conn) {
	defer clientCon.Close()

	fmt.Println("Client connnected")
	// Create a buffer for incoming data.
	buf := make([]byte, 4096)

	for {
		// Read data from the connection.
		n, err := clientCon.Read(buf)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Connection terminated")
			break
		}

		fmt.Println(buf[:n])
	}
}
