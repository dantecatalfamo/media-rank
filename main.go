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
			log.Println(err)
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

		row := server.db.QueryRow("SELECT COUNT(*) FROM media WHERE deleted = true")
		if row.Err() != nil {
			log.Fatalf("main failed to get number of deleted files: %s", err)
			os.Exit(1)
		}
		var deletedFiles int
		if err := row.Scan(&deletedFiles); err != nil {
			log.Fatalf("main failed to scan deleted files: %s", err)
			os.Exit(1)
		}
		if deletedFiles > 0 {
			log.Printf("Removing %d deleted files from database", deletedFiles)
			// Delete media marked by scanMedia
			if _, err := server.db.Exec("DELETE FROM media WHERE deleted = true"); err != nil {
				log.Fatalf("main failed to remove deleted media: %s", err)
				os.Exit(1)
			}
		}
	}()

	log.Println("setting up routes")
	SetupRoutes(server)

	log.Printf("Starting server on http://%s\n", address)
	err = http.ListenAndServe(*userAddress, nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %s\n", err)
	}
}
