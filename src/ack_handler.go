package main

import (
	"bytes"
	"fmt"
	"net/http"
)

var (
	receivedAck bool
)

func sendAcknowledgement(port int) {
	url := fmt.Sprintf("%s%d:%d/ack", baseUrl, port, port)
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
