package main

import (
	"fmt"
	"net/http"
	"strconv"
)

var (
	baseUrl = "http://app-"
	repos   = []string{}
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
	fmt.Println("Listening on port", nodePort)
	if err := http.ListenAndServe(":"+strconv.Itoa(nodePort), nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
