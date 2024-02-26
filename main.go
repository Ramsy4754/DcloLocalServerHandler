package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	serverCmd      *exec.Cmd
	serverCmdMutex sync.Mutex
)

func main() {
	http.HandleFunc("/start", startServerHandler)
	http.HandleFunc("/stop", stopServerHandler)

	fmt.Println("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startServerHandler(w http.ResponseWriter, r *http.Request) {
	serverCmdMutex.Lock()
	defer serverCmdMutex.Unlock()

	if serverCmd != nil {
		_, _ = fmt.Fprintf(w, "Server is already running.\n")
		return
	}

	envVars, err, done := getEnvVars(w)
	if done {
		return
	}

	var outBuffer, errBuffer bytes.Buffer

	serverCmd = exec.Command("D:\\00_Development\\Dclo_Back\\venv\\Scripts\\python.exe", "D:\\00_Development\\Dclo_Back\\run.py")
	serverCmd.Env = append(os.Environ(), envVars...)
	serverCmd.Env = append(serverCmd.Env, "SYSTEM_TYPE=LOCAL")

	serverCmd.Stdout = &outBuffer
	serverCmd.Stderr = &errBuffer

	err = serverCmd.Start()
	if err != nil {
		http.Error(w, "Failed to run python server", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Starting server...\nOutput:\n%s\nError:\n%s", outBuffer.String(), errBuffer.String())
}

func getEnvVars(w http.ResponseWriter) ([]string, error, bool) {
	envVars, err := readEnvFile("D:\\00_Development\\Dclo_Back\\.env")
	if err != nil {
		http.Error(w, "Failed to read .env file", http.StatusInternalServerError)
		return nil, nil, true
	}
	return envVars, err, false
}

func readEnvFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

func stopServerHandler(w http.ResponseWriter, r *http.Request) {
	serverCmdMutex.Lock()
	defer serverCmdMutex.Unlock()

	if serverCmd != nil && serverCmd.Process != nil {
		err := serverCmd.Process.Kill()
		if err != nil {
			http.Error(w, "Failed to stop server", http.StatusInternalServerError)
			return
		}
		serverCmd = nil
		fmt.Fprintf(w, "Server stopped successfully.")
	} else {
		fmt.Fprintf(w, "Server is not running.")
	}
}
