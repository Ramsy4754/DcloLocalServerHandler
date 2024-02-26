package main

import (
	"bufio"
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
	logFile, err := os.OpenFile("D:/03_Private/Go_Local_Server/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set the output of the log package to the log file
	log.SetOutput(logFile)

	http.HandleFunc("/start", startServerHandler)
	http.HandleFunc("/stop", stopServerHandler)

	log.Println("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startServerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("startServerHandler", r.RemoteAddr)
	serverCmdMutex.Lock()
	defer serverCmdMutex.Unlock()

	if serverCmd != nil && serverCmd.Process != nil {
		_, _ = fmt.Fprintf(w, "Server is already running.\n")
		return
	}

	envVars, err, done := getEnvVars(w)
	if done {
		return
	}

	serverCmd = exec.Command("D:/00_Development/Dclo_Back/venv/Scripts/python.exe", "D:/00_Development/Dclo_Back/run.py")
	serverCmd.Env = append(os.Environ(), envVars...)
	serverCmd.Env = append(serverCmd.Env, "SYSTEM_TYPE=LOCAL")

	err = serverCmd.Start()
	if err != nil {
		http.Error(w, "Failed to run python server", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Starting server...")
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
	log.Println("stopServerHandler", r.RemoteAddr)
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
