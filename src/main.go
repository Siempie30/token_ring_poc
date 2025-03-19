package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func Init() error {
	err := InitRing()
	if err != nil {
		fmt.Println("Error initializing ring:", err)
		return err
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
	http.HandleFunc("/addition", handlePortAddition)
	http.HandleFunc("/removal", handlePortRemoval)
	fmt.Println("Listening on port", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
