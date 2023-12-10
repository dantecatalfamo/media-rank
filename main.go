package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	address = "127.0.0.1:4400"
	dbName = "media.db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <media directory>\n", os.Args[0])
		os.Exit(1)
	}

	mediaDirectory := os.Args[1]

	if err := os.Chdir(mediaDirectory); err != nil {
		log.Fatalf("failed to change to media directory: %s", err)
	}

	server, err := NewServer(dbName)
	if err != nil {
		log.Fatalf("creating new server: %s", err)
	}

	log.Println("beginning media scan")
	go scanMedia(server, ".")

	log.Println("setting up routes")
	SetupRoutes(server)

	log.Printf("Starting server on http://%s\n", address)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %s\n", err)
	}
}
