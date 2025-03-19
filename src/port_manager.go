package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	portFile = "ring_ports.txt"
	port     = -1
)

func getNextPort(currentPort int) (int, error) {
	file, err := os.Open(portFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ports []string
	for scanner.Scan() {
		ports = append(ports, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	for i, p := range ports {
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

	if err := os.Rename(tempFile, portFile); err != nil {
		return err
	}

	fmt.Println("Removed port", port)
	return nil
}

func sendPortRemoval(port int) error {
	file, err := os.Open(portFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		fmt.Println("Sending port removal request to", line)

		targetPort, _ := strconv.Atoi(line)
		url := fmt.Sprintf("http://localhost:%d/removal", targetPort)
		if targetPort == port {
			continue
		}
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

	body, err := io.ReadAll(r.Body)
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
