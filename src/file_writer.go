package main

import (
	"fmt"
	"os"
	"time"
)

func writeToFile(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("Written by %s at %s\n", os.Getenv("PORT"), timestamp)
	if _, err := f.WriteString(entry); err != nil {
		fmt.Println("Error writing to file:", err)
	}
	fmt.Println("Writing to file")

	time.Sleep(2 * time.Second)
}
