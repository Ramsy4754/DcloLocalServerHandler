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
	dcloServerCmd      *exec.Cmd
	dcloServerCmdMutex sync.Mutex
)

var (
	adminServerCmd      *exec.Cmd
	adminServerCmdMutex sync.Mutex
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

	http.HandleFunc("/start", startDcloServerHandler)
	http.HandleFunc("/stop", stopDcloServerHandler)

	http.HandleFunc("/start-admin", startAdminServerHandler)
	http.HandleFunc("/stop-admin", stopAdminServerHandler)

	log.Println("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startAdminServerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("startAdminServerHandler", r.RemoteAddr)
	adminServerCmdMutex.Lock()
	defer adminServerCmdMutex.Unlock()

	if adminServerCmd != nil && adminServerCmd.Process != nil {
		_, _ = fmt.Fprintf(w, "Server is already running.\n")
		return
	}

	adminServerCmd = exec.Command("D:/00_Development/Dclo_Admin/.venv/Scripts/python.exe", "D:/00_Development/Dclo_Admin/run.py", "--host", "0.0.0.0")

	// Set Environment Variables
	adminServerCmd.Env = append(os.Environ(), "SYSTEM_TYPE=local")

	err := adminServerCmd.Start()
	if err != nil {
		http.Error(w, "Failed to run python server", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Starting server...")
}

func stopAdminServerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("stopAdminServerHandler", r.RemoteAddr)
	adminServerCmdMutex.Lock()
	defer adminServerCmdMutex.Unlock()

	if adminServerCmd != nil && adminServerCmd.Process != nil {
		err := adminServerCmd.Process.Kill()
		if err != nil {
			http.Error(w, "Failed to stop server", http.StatusInternalServerError)
			return
		}
		adminServerCmd = nil
		_, _ = fmt.Fprintf(w, "Server stopped successfully.")
	} else {
		_, _ = fmt.Fprintf(w, "Server is not running.")
	}
}

func startDcloServerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("startDcloServerHandler", r.RemoteAddr)
	dcloServerCmdMutex.Lock()
	defer dcloServerCmdMutex.Unlock()

	if dcloServerCmd != nil && dcloServerCmd.Process != nil {
		_, _ = fmt.Fprintf(w, "Server is already running.\n")
		return
	}

	envVars, err, done := getEnvVars(w)
	if done {
		return
	}

	dcloServerCmd = exec.Command("poetry", "run", "flask", "run")
	dcloServerCmd.Dir = "D:/00_Development/Dclo_Back"
	dcloServerCmd.Env = append(os.Environ(), envVars...)
	//dcloServerCmd.Env = append(dcloServerCmd.Env, "SYSTEM_TYPE=LOCAL")
	dcloServerCmd.Env = append(dcloServerCmd.Env, "FLASK_APP=run")

	// Run redis server
	cmd := exec.Command("wsl", "redis-server")
	err = cmd.Start()
	if err != nil {
		log.Printf("Failed to start redis-server via WSL: %v", err)
		return
	}

	err = dcloServerCmd.Start()
	if err != nil {
		http.Error(w, "Failed to run python server", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Starting server...")
}

func getEnvVars(w http.ResponseWriter) ([]string, error, bool) {
	envVars, err := readEnvFile("D:/00_Development/Dclo_Back/.env")
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

func stopDcloServerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("stopDcloServerHandler", r.RemoteAddr)
	dcloServerCmdMutex.Lock()
	defer dcloServerCmdMutex.Unlock()

	if dcloServerCmd != nil && dcloServerCmd.Process != nil {
		err := dcloServerCmd.Process.Kill()
		if err != nil {
			http.Error(w, "Failed to stop server", http.StatusInternalServerError)
			return
		}
		dcloServerCmd = nil
		fmt.Fprintf(w, "Server stopped successfully.")
	} else {
		fmt.Fprintf(w, "Server is not running.")
	}
}
