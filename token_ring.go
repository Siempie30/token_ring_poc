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

func getPreviousPort(currentPort int) (int, error) {
	data, err := ioutil.ReadFile(portFile)
	if err != nil {
		return 0, err
	}
	ports := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i, p := range ports {
		p = strings.TrimSpace(p)
		if p == strconv.Itoa(currentPort) {
			if i == 0 {
				return strconv.Atoi(ports[len(ports)-1])
			}
			return strconv.Atoi(ports[i-1])
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
	fmt.Println("Writing to file")

	time.Sleep(2 * time.Second)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Send ack to node that sent the token
	previousPort := r.Header.Get("From-Port")
	if previousPort != "" {
		prevPort, err := strconv.Atoi(previousPort)
		if err == nil {
			sendAcknowledgement(prevPort)
		}
	}
	// Access the shared resource, now that we have the token
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

func sendAcknowledgement(port int) {
	url := fmt.Sprintf("http://localhost:%d/ack", port)
	_, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte("ack")))
	if err != nil {
		fmt.Println("Error sending acknowledgement to", port, ":", err)
		return
	}
	fmt.Println("Sent acknowledgement to", port)
}

func handleAcknowledgement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Acknowledgement received")
}

func postToken(port int) {
	url := fmt.Sprintf("http://localhost:%d/token", port)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte("token")))
	if err != nil {
		fmt.Println("Error creating token request:", err)
		return
	}
	req.Header.Set("From-Port", os.Getenv("PORT"))
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending token to", port, ":", err)
	}
	fmt.Println("Posted token to", port)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT environment variable not set")
		return
	}

	http.HandleFunc("/token", handleToken)
	http.HandleFunc("/ack", handleAcknowledgement)
	fmt.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
