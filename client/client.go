package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: client <PORT>")
		os.Exit(1)
	}
	localPort := os.Args[1]

	// 1. Connect to control port
	controlConn, err := net.Dial("tcp", "localhost:7835")
	if err != nil {
		panic(err)
	}

	// 2. Send EXPOSE command over the control connection
	fmt.Fprintf(controlConn, "EXPOSE %s\n", localPort)

	// 3. Read OK response with public port
	reader := bufio.NewReader(controlConn)
	line, _ := reader.ReadString('\n')
	fmt.Print("Server response: ", line)

	//4. Listen for CONNECTION <id> [incoming user connections (from server)] messages
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

			// connect to data port and claim ID
			go handleTunnel(connID, localPort)
		}
	}
}

func handleTunnel(connID, localPort string) {
	dataConn, err := net.Dial("tcp", "localhost:7836")
	if err != nil {
		fmt.Println("Error connecting to data tunnel:", err)
		return
	}
	fmt.Fprintf(dataConn, "ID %s\n", connID)

	localConn, err := net.Dial("tcp", "localhost:"+localPort)
	if err != nil {
		fmt.Println("Error connecting to local service:", err)
		dataConn.Close()
		return
	}

	//This pipes both ways
	go io.Copy(dataConn, localConn)
	go io.Copy(localConn, dataConn)
}
