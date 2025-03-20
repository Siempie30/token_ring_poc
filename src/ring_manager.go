package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	portDir      = "ring"
	portFileBase = "ring_ports.txt"
	port         = -1
)

func InitRing() error {
	// Get port from environment variable
	portString := os.Getenv("PORT")
	if portString == "" {
		fmt.Println("PORT environment variable not set")
		return fmt.Errorf("PORT environment variable not set")
	}
	var err error
	port, err = strconv.Atoi(portString)
	if err != nil {
		fmt.Println("Invalid port:", err)
		return err
	}

	// Get repos from environment variable
	repoList := os.Getenv("REPOS")
	if repoList == "" {
		fmt.Println("REPOS environment variable not set")
		return fmt.Errorf("REPOS environment variable not set")
	}
	repos = strings.Split(repoList, ",")
	for _, repo := range repos {
		fmt.Println("Repo:", repo)
	}

	// Check if ring file exists for each repo
	for _, repo := range repos {
		ringFile := getRingFile(repo)
		if _, err := os.Stat(ringFile); os.IsNotExist(err) {
			fmt.Println("Ring file does not exist for", repo)
			return fmt.Errorf("ring file does not exist for %s", repo)
		}
	}

	for _, repo := range repos {
		// Check if port is in ring
		if !isInRing(port, repo) {
			// Send message to gateways in ring to append this gateway to the ring
			err = sendPortAddition(port, repo)
			if err != nil {
				fmt.Println("Error sending port addition request:", err)
				return err
			}
			// Append own port to ring
			appendPort(port, repo)
		} else {
			// This gateway is already part of the ring, so simply start waiting for token
		}
	}

	return nil
}

func getRingFile(repo string) string {
	return fmt.Sprintf("%s/%s-%s", portDir, repo, portFileBase)
}

func isInRing(port int, repo string) bool {
	file, err := os.Open(getRingFile(repo))
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strconv.Itoa(port) {
			return true
		}
	}

	return false
}

func getNextPort(currentPort int, repo string) (int, error) {
	file, err := os.Open(getRingFile(repo))
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

// Append a port to the ring file. Currently assumes that the ring file ends with a newline.
func appendPort(port int, repo string) error {
	file, err := os.OpenFile(getRingFile(repo), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(port) + "\n")
	if err != nil {
		return err
	}

	fmt.Println("Added port", port)
	return nil
}

func removePort(port int, repo string) error {
	tempFile := getRingFile(repo) + ".tmp"
	inputFile, err := os.Open(getRingFile(repo))
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

	if err := os.Rename(tempFile, getRingFile(repo)); err != nil {
		return err
	}

	fmt.Println("Removed port", port)
	return nil
}

func sendPortRemoval(repo string, port int) error {
	file, err := os.Open(getRingFile(repo))
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
		url := fmt.Sprintf("%s%d:%d/removal", baseUrl, targetPort, targetPort)
		payload := fmt.Sprintf("%d,%s", port, repo)
		_, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte(payload)))
		if err != nil {
			fmt.Println("Error sending port removal request to", targetPort, ":", err)
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

	data := strings.SplitN(strings.TrimSpace(string(body)), ",", 2)
	if len(data) != 2 {
		http.Error(w, "Invalid payload format", http.StatusBadRequest)
		return
	}

	removedPort, err := strconv.Atoi(data[0])
	if err != nil {
		http.Error(w, "Invalid port format", http.StatusBadRequest)
		return
	}

	repo := data[1]
	if !slices.Contains(repos, repo) {
		fmt.Println("Received request for removal of port from invalid repo, ignoring")
		return
	}

	err = removePort(removedPort, repo)
	if err != nil {
		fmt.Println("Failed to remove port", removedPort, "from repo", repo, ":", err)
		return
	}
}

func sendPortAddition(port int, repo string) error {
	file, err := os.Open(getRingFile(repo))
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
		fmt.Println("Sending port addition request to", line)

		targetPort, _ := strconv.Atoi(line)
		url := fmt.Sprintf("%s%d:%d/addition", baseUrl, targetPort, targetPort)
		payload := fmt.Sprintf("%d,%s", port, repo)
		_, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte(payload)))
		if err != nil {
			fmt.Println("Error sending port addition request to", targetPort, ":", err)
			return err
		}
	}
	return nil
}

func handlePortAddition(w http.ResponseWriter, r *http.Request) {
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

	data := strings.SplitN(strings.TrimSpace(string(body)), ",", 2)
	if len(data) != 2 {
		http.Error(w, "Invalid payload format", http.StatusBadRequest)
		return
	}

	addedPort, err := strconv.Atoi(data[0])
	if err != nil {
		http.Error(w, "Invalid port format", http.StatusBadRequest)
		return
	}

	repo := data[1]

	err = appendPort(addedPort, repo)
	if err != nil {
		fmt.Println("Failed to append port", addedPort, "to repo", repo, ":", err)
		return
	}
}
