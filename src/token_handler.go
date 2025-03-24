package main

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"
)

var (
	receivedTokenChan = make(chan bool, 1)
)

func postToken(repo string, targetPort int) {
	url := fmt.Sprintf("%s%d:%d/token", baseUrl, targetPort, targetPort)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(repo)))
	if err != nil {
		fmt.Println("Error creating token request:", err)
		return
	}
	req.Header.Set("From-Port", strconv.Itoa(nodePort))
	receivedAck = false
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending token to", targetPort, ":", err)
	}

	for start := time.Now(); time.Since(start) < 5*time.Second; {
		if receivedAck {
			return
		}
	}
	nextPort, _ := getNextPort(targetPort, repo)
	sendPortRemoval(repo, targetPort)
	postToken(repo, nextPort)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Determine the port that sent the token
	previousPort := r.Header.Get("From-Port")
	prevPort := -1
	if previousPort != "" {
		var err error
		prevPort, err = strconv.Atoi(previousPort)
		if err != nil {
			fmt.Println("Invalid port:", err)
			return
		}
		receivedTokenChan <- true
		fmt.Println("Received token from", prevPort)
	} else {
		fmt.Println("Received token from unknown port")
	}

	// Read repo name from body
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
	if !slices.Contains(repos, reponame) {
		fmt.Println("Received token for unknown repository", reponame, "ignorging")
		return
	}
	if prevPort != -1 {
		sendAcknowledgement(prevPort, reponame)
	}

	filename := "output/" + reponame + "_common.txt"
	writeToFile(filename)

	targetPort, err := getNextPort(nodePort, reponame)
	if err != nil {
		fmt.Println("Error getting next port:", err)
		return
	}

	postToken(reponame, targetPort)
}
