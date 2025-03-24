package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

var (
	receivedAck bool
)

func sendAcknowledgement(targetPort int, repo string) {
	url := fmt.Sprintf("%s%d:%d/ack", baseUrl, targetPort, targetPort)
	_, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte(repo)))
	if err != nil {
		fmt.Println("Error sending acknowledgement to", targetPort, ":", err)
		return
	}
	fmt.Println("Sent acknowledgement to", targetPort)
}

func handleAcknowledgement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Acknowledgement received")
	receivedAck = true

	// Read repo name from body
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	repo := buf.String()

	// Determine the maximum time for the cycle
	t_node := maxWriteTimeMs // Node time is now assumed ot be the same as the maximum write time
	ringsize, err := getRingSize(repo)
	t_cycle := ringsize * t_node
	t_cycle_ms := time.Duration(t_cycle) * time.Millisecond
	if err != nil {
		fmt.Println("Failed to get ring size", err)
		return
	}

	go func() {
		// Wait for the cycle to end
		select {
		case <-receivedTokenChan:
			fmt.Println("Received token in time!")
		case <-time.After(t_cycle_ms):
			fmt.Println("Token not received in time! (is token owner dead?)")
			// Post the token to the next port
			nextPort, err := getNextPort(nodePort, repo)
			if err != nil {
				fmt.Println("Failed to get next port", err)
				return
			}
			fmt.Println("Generating new token and posting to ", nextPort)
			postToken(repo, nextPort)
		}
	}()
}
