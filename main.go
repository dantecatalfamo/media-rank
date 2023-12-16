package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
)

const (
	address = "127.0.0.1:4400"
	dbName = "media.db"
)

func main() {
	userAddress := flag.String("addr", address, "address:port to start the server on")
	mediaDirectory := flag.String("media", ".", "location of media directory")
	flag.Parse()

	if err := os.Chdir(*mediaDirectory); err != nil {
		log.Fatalf("failed to change to media directory: %s", err)
	}

	server, err := NewServer(dbName)
	if err != nil {
		log.Fatalf("creating new server: %s", err)
	}

	ctx := context.Background()
	log.Println("beginning media scan")
	scanMedia(ctx, server, ".")

	log.Println("setting up routes")
	SetupRoutes(server)

	log.Printf("Starting server on http://%s\n", address)
	err = http.ListenAndServe(*userAddress, nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %s\n", err)
	}
}
