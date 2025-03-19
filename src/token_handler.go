package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

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

	for start := time.Now(); time.Since(start) < 5*time.Second; {
		if receivedAck {
			return
		}
	}
	nextPort, _ := getNextPort(port)
	sendPortRemoval(port)
	postToken(nextPort)
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
