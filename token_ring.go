package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	commonFile  = "../common.txt"
	portFile    = "ring_ports.txt"
	receivedAck = false
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

func removePort(port int) error {
	tempFile := portFile + ".tmp"
	inputFile, err := os.Open(portFile)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)

	for scanner.Scan() {
		line := scanner.Text()
		// Copy the string into the temp file if it does not contain the port that should be removed
		if !strings.Contains(line, strconv.Itoa(port)) {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
	}

	writer.Flush()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Replace original file with temp file
	if err := os.Rename(tempFile, portFile); err != nil {
		return err
	}

	fmt.Println("Removed port", port)
	return nil
}

func writeToFile() {
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

	// Simulate procesing time
	time.Sleep(2 * time.Second)
}

func postToken(port int) {
	url := fmt.Sprintf("http://localhost:%d/token", port)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte("token")))
	if err != nil {
		fmt.Println("Error creating token request:", err)
		return
	}
	fmt.Println("Posting token to", port)
	req.Header.Set("From-Port", os.Getenv("PORT"))
	receivedAck = false
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending token to", port, ":", err)
	}

	// Wait for acknowledgement for five seconds
	for start := time.Now(); time.Since(start) < 5*time.Second; {
		if receivedAck == true {
			// Next node confirmed having received the token
			return
		}
	}
	// Acknowledgement not received, so assume next node is down. Remove the failing port and attempt the first next port
	nextPort, _ := getNextPort(port)
	sendPortRemoval(port)
	postToken(nextPort)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Send ack to token sender
	previousPort := r.Header.Get("From-Port")
	if previousPort != "" {
		prevPort, err := strconv.Atoi(previousPort)
		if err == nil {
			fmt.Println("Received token from", prevPort)
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
	receivedAck = true
}

func sendPortRemoval(port int) error {
	file, err := os.Open(portFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy lines into buffer first
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		fmt.Println("Sending port removal request to", line)

		targetPort, _ := strconv.Atoi(line)
		url := fmt.Sprintf("http://localhost:%d/removal", targetPort)
		if targetPort == port { // Skip the port that is to be removed
			continue
		}
		// Send message that port is to be removed
		_, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte(strconv.Itoa(port))))
		if err != nil {
			fmt.Println("Error sending port removal request to", targetPort, ":", err)
			return err
		}
	}
	return nil
}

func handlePortRemoval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	removedPort, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		http.Error(w, "Invalid port format", http.StatusBadRequest)
		return
	}

	removePort(removedPort)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT environment variable not set")
		return
	}

	http.HandleFunc("/token", handleToken)
	http.HandleFunc("/ack", handleAcknowledgement)
	http.HandleFunc("/removal", handlePortRemoval)
	fmt.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
