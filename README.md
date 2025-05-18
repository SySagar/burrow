# 🕳️ Go TCP Tunneling (Like Ngrok)

[![Status - Development](https://img.shields.io/badge/Status-Development-FDDA0D)](https://)

A lightweight reverse TCP tunnel server + client in Go — like `ngrok`  
Expose your local services running on `localhost` to the public internet via a single public server.

## ✨ Features

- Single binary client and server
- Reverse tunneling with no authentication
- Simple protocol using TCP

## 📦 Architecture

```txt
   ┌────────────┐                      ┌──────────────┐
   │  Client    │     control (7835)   │   Server     │
   │ (localhost)│◀────────────────────▶│ (public VPS) │
   └────┬───────┘                      └────┬─────────┘
        │                                   │
        │  data tunnel (7836)               │
        │◀──────────────────────────────────┤
        │                                   │
   ┌────▼───────┐                     ┌─────▼────────────┐
   │ Local App  │◀─── localhost:5000 ─│ External Request │
   │  (e.g. API)│                     │  (e.g. curl)     │
   └────────────┘                     └──────────────────┘
```

# Key Elements
## 🔌 Control Connection (:7835)

This connection is used for the initial setup and communication of connection requests.

**Client Initiates:**

1.  The client sends an `EXPOSE` command followed by the local port they want to expose.

    ```
    EXPOSE 5000
    ```

**Server Responds:**

1.  Upon receiving the `EXPOSE` command, the server acknowledges with an `OK` message and provides a publicly accessible port (`<publicPort>`).

    ```
    OK 12345
    ```

**Server Notifies of Incoming Connection:**

1.  When an external user connects to the `<publicPort>` on the server, the server sends a `CONNECTION` message to the client, including a unique connection identifier (`<id>`).

    ```
    CONNECTION abcdef123
    ```

## 🔄 Data Tunnel Connection (:7836)

This connection is established to create a tunnel for the actual data transfer.

**Client Connects Back:**

1.  Upon receiving the `CONNECTION` message, the client establishes a new connection to the server on a different port (:7836).

2.  Immediately after connecting, the client sends an `ID` message containing the `<id>` received in the `CONNECTION` message.

    ```
    ID abcdef123
    ```

**Server Pipes Connections:**

1.  The server receives the `ID` and identifies the waiting user connection associated with that `<id>`.

2.  The server then establishes a direct pipe between the client's Data Tunnel Connection and the waiting user's connection. Any data sent on one connection is directly forwarded to the other, creating a tunnel.

This two-step process allows the server to manage multiple exposed ports and efficiently connect incoming requests to the appropriate client.


## 📦 Getting Started

Follow these steps to set up and use the TCP tunnel.

---

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/go-tcp-tunnel.git
cd go-tcp-tunnel
```

### 2. Run a Local Service (for Testing)

```bash
python3 -m http.server 5000
```

### 3.  Start the Tunnel Server (for Testing)

```bash
go run server/main.go
```

### 4.  Start the Tunnel Client (on Your Local Machine)

```bash
go run client/main.go
```
If successful, you’ll see responses like:


Server response: OK 9090

Server response: CONNECTION 12345

### 5.  Test it

```bash
curl http://localhost:9090
```
