package backend

import (
    "fmt"
    "net"
)

// StartPeerServer starts a TCP server for peer-to-peer connections
func StartPeerServer(port string) {
    listener, err := net.Listen("tcp", ":"+port)
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer listener.Close()
    fmt.Println("Peer server listening on port", port)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Connection error:", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    fmt.Println("New peer connected:", conn.RemoteAddr())
}