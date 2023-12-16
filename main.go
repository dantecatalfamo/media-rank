package main

import (
	"context"
	"flag"
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
	log.Printf("beginning media scan of %s\n", *mediaDirectory)
	errChan, finishChan := scanMedia(ctx, server, ".")
	go func() {
		for err := range(errChan) {
			fmt.Println(err)
		}
	}()
	go func() {
		var finished uint = 0
		for finish := range(finishChan) {
			if finish {
				fmt.Printf(".")
			} else {
				fmt.Printf("!")
			}
			finished++
			if (finished % 100) == 0 {
				fmt.Printf("\n{%d}", finished)
			}
		}
		fmt.Printf("\nfinished media scan, total: %d\n", finished)
	}()

	log.Println("setting up routes")
	SetupRoutes(server)

	log.Printf("Starting server on http://%s\n", address)
	err = http.ListenAndServe(*userAddress, nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %s\n", err)
	}
}
