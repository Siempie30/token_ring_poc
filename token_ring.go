package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	commonFile = "common.txt"
	portFile   = "ring_ports.txt"
	mutex      sync.Mutex
)

func getNextPort(currentPort int) (int, error) {
	data, err := ioutil.ReadFile(portFile)
	if err != nil {
		return 0, err
	}
	ports := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i, p := range ports {
		p = strings.TrimSpace(p)
		if p == strconv.Itoa(currentPort) {
			return strconv.Atoi(ports[(i+1)%len(ports)])
		}
	}
	return 0, fmt.Errorf("current port not found in ring")
}

func writeToFile() {
	mutex.Lock()
	defer mutex.Unlock()

	f, err := os.OpenFile(commonFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("Written by %s at %s\n", os.Getenv("PORT"), timestamp)
	if _, err := f.WriteString(entry); err != nil {
		fmt.Println("Error writing to file:", err)
	}

	time.Sleep(2 * time.Second)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	writeToFile()

	currentPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("Invalid port:", err)
		return
	}

	nextPort, err := getNextPort(currentPort)
	if err != nil {
		fmt.Println("Error getting next port:", err)
		return
	}

	postToken(nextPort)
}

func postToken(port int) {
	url := fmt.Sprintf("http://localhost:%d/token", port)
	_, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte("token")))
	if err != nil {
		fmt.Println("Error sending token to", port, ":", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT environment variable not set")
		return
	}

	http.HandleFunc("/token", handleToken)
	fmt.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
