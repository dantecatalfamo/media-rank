package main

import (
	"log"
	"net/http"
	"os"
	"path"
)

const (
	address = "127.0.0.1:4400"
	dbName = "media.db"
)

func main() {
	mediaDirectory := os.Args[1]
	dbPath := path.Join(mediaDirectory, dbName)

	server, err := NewServer(dbPath)
	if err != nil {
		log.Fatalf("creating new server: %s", err)
	}
	log.Print(server)
	scanMedia(server, mediaDirectory)
	log.Printf("Starting server on http://%s\n", address)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %s\n", err)
	}
}
