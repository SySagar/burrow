package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

func main() {
	// 1. Connect to control port
	controlConn, err := net.Dial("tcp", "localhost:7835")
	if err != nil {
		panic(err)
	}

	// 2. Send EXPOSE command over the control connection
	fmt.Fprintf(controlConn, "EXPOSE 5000\n")

	// 3. Read OK response with public port
	reader := bufio.NewReader(controlConn)
	line, _ := reader.ReadString('\n')
	fmt.Print("Server response: ", line)

	// 4. Listen for CONNECTION <id> [incoming user connections (from server)] messages
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading server response:", err)
			return
		}
		fmt.Print("Server response: ", line)
		if strings.HasPrefix(line, "CONNECTION") {
			connID := strings.TrimSpace(strings.Split(line, " ")[1])
			fmt.Println("Got connection ID:", connID)

			// 5. Connect to data port and claim the ID
			go handleTunnel(connID)
		}
	}
}

func handleTunnel(connID string) {
	// Connect to the tunnel data port(7836), which is waiting for tunnel connection
	dataConn, err := net.Dial("tcp", "localhost:7836")
	if err != nil {
		fmt.Println("Error connecting to data tunnel:", err)
		return
	}
	fmt.Fprintf(dataConn, "ID %s\n", connID)

	// Connect to local service (port 5000)
	localConn, err := net.Dial("tcp", "localhost:5000")
	if err != nil {
		fmt.Println("Error connecting to local service:", err)
		dataConn.Close()
		return
	}

	// Pipe both ways
	go io.Copy(dataConn, localConn)
	go io.Copy(localConn, dataConn)
}
