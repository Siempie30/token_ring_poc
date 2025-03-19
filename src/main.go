package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func Init() error {
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

	if _, err := os.Stat(portFile); os.IsNotExist(err) {
		return fmt.Errorf("ring file does not exist")
	}

	return nil
}

func main() {
	err := Init()
	if err != nil {
		fmt.Println("Error initializing:", err)
		return
	}

	http.HandleFunc("/token", handleToken)
	http.HandleFunc("/ack", handleAcknowledgement)
	http.HandleFunc("/removal", handlePortRemoval)
	fmt.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
