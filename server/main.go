package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// defining a tunnel strut to hold the state
type Tunnel struct {
	controlConn net.Conn // control connection with client
	localPort   string   // local port client wants to expose
}

var (
	tunnels = make(map[string]*Tunnel) // publicPort -> Tunnel
	mu      sync.Mutex                 // a mutex mu of type sync.Mutex
)

// temporarily store incoming tunnel connection IDs and pair them with the external connections that are waiting
var (
	pendingTunnels = make(map[string]net.Conn)
	pendingMu      sync.Mutex
)

func main() {
	// control server on TCP port 7835
	go func() {
		listener, err := net.Listen("tcp", ":7835")

		if err != nil {
			panic(err)
		}

		fmt.Println("Server started on port 7835 (control port)")

		//infinte loop
		for {
			conn, err := listener.Accept()

			if err != nil {
				continue
			}
			go handleControlConnection(conn)
		}
	}()

	// Start data tunnel listener
	// Data tunnel port (client sends ID <id> to match the incoming connection)
	go func() {
		tunnelListener, err := net.Listen("tcp", ":7836")
		if err != nil {
			panic(err)
		}
		fmt.Println("Server started on port 7836 (data tunnel port)")
		for {
			conn, err := tunnelListener.Accept()
			if err != nil {
				continue
			}
			go handleClientDataConnection(conn)
		}
	}()

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("➡️ Received request at /")
	// 	fmt.Fprintln(w, "Hello from your local web server!")
	// })

	// keep main thread running forever
	select {}

}

func handleControlConnection(conn net.Conn) {
	// defer conn.Close()

	reader := bufio.NewReader(conn)

	// Step 1: Expect a message like "EXPOSE 5000\n"
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from control connection:", err)
		return
	}

	parts := strings.Fields(line)
	if len(parts) != 2 || parts[0] != "EXPOSE" {
		fmt.Fprintln(conn, "ERROR invalid EXPOSE command")
		return
	}
	localPort := parts[1]

	// Seeds the random number generator using the current time.
	rand.Seed(time.Now().UnixNano())

	// Choose a random public port (9000–9100)
	var publicPort int
	for {
		publicPort = rand.Intn(100) + 9000
		mu.Lock()
		if _, exists := tunnels[strconv.Itoa(publicPort)]; !exists {
			break
		}
		mu.Unlock()
	}

	// Start listening on the public port (for ex 9090)
	pubListener, err := net.Listen("tcp", fmt.Sprintf(":%d", publicPort))
	if err != nil {
		fmt.Fprintln(conn, "ERROR failed to bind public port")
		mu.Unlock()
		return
	}

	fmt.Printf("Exposing client local port %s at public port %d\n", localPort, publicPort)

	// Store the new tunnel object in map
	tunnel := &Tunnel{
		controlConn: conn,
		localPort:   localPort,
	}
	tunnels[strconv.Itoa(publicPort)] = tunnel
	mu.Unlock()

	// Tell client what public port to use
	fmt.Fprintf(conn, "OK %d\n", publicPort)

	//Handle incoming connections on the public port
	go func() {
		for {
			externalConn, err := pubListener.Accept()
			if err != nil {
				fmt.Println("Error accepting external connection:", err)
				return
			}

			// Generate a short random ID for the tunnel session
			connID := fmt.Sprintf("%d", rand.Intn(100000))

			// Notify client via control connection
			fmt.Fprintf(conn, "CONNECTION %s\n", connID)

			// Wait for client to dial back to handle this ID
			go waitForClientTunnel(publicPort, connID, externalConn)
		}
	}()

	// Block forever to keep controlConn alive
	select {} // This keeps the goroutine alive and controlConn open
}

// timeout after 10s from client
func waitForClientTunnel(publicPort int, connID string, externalConn net.Conn) {

	// Save external connection by ID into map
	pendingMu.Lock()
	pendingTunnels[connID] = externalConn
	pendingMu.Unlock()

	// We'll wait up to 10 seconds for the client to connect back
	time.AfterFunc(10*time.Second, func() {
		pendingMu.Lock()

		if conn, ok := pendingTunnels[connID]; ok {
			conn.Close()
			delete(pendingTunnels, connID)
			fmt.Println("Client did not connect back in time, closed external connection.")
		}
		pendingMu.Unlock()

	})

}

//	reverse tunnel setup: Handles when the client connects back to the server on port 7836
//
// CLIENT SENDS : ID abc123\n
func handleClientDataConnection(conn net.Conn) {
	// defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read from client tunnel connection")
		return
	}

	parts := strings.Fields(line)
	if len(parts) != 2 || parts[0] != "ID" {
		fmt.Fprintln(conn, "ERROR invalid ID command")
		return
	}
	connID := parts[1]

	// Look up the external connection waiting for this tunnel
	pendingMu.Lock()
	externalConn, ok := pendingTunnels[connID]
	if ok {
		delete(pendingTunnels, connID)
	}
	pendingMu.Unlock()

	if !ok {
		fmt.Fprintln(conn, "ERROR unknown connection ID")
		return
	}

	fmt.Printf("Piping tunnel %s to external connection\n", connID)

	// Pipe data both ways, Streams TCP traffic both ways – this forms the actual tunnel
	go io.Copy(externalConn, conn)
	go io.Copy(conn, externalConn)
}
