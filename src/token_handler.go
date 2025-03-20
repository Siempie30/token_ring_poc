package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func postToken(repo string, port int) {
	url := fmt.Sprintf("%s%d:%d/token", baseUrl, port, port)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(repo)))
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

	for start := time.Now(); time.Since(start) < 5*time.Second; {
		if receivedAck {
			return
		}
	}
	nextPort, _ := getNextPort(port)
	sendPortRemoval(repo, port)
	postToken(repo, nextPort)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	previousPort := r.Header.Get("From-Port")
	if previousPort != "" {
		prevPort, err := strconv.Atoi(previousPort)
		if err == nil {
			fmt.Println("Received token from", prevPort)
			sendAcknowledgement(prevPort)
		}
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	reponame := buf.String()
	if reponame == "" {
		fmt.Println("Received empty token")
		return
	}

	filename := "output/" + reponame + "_common.txt"

	writeToFile(filename)

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

	postToken(reponame, nextPort)
}
