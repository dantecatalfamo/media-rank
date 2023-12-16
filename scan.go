package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var acceptedFileTypes = []string{
	"jpg",
	"jpeg",
	"png",
	"gif",
	// "mp4",
	// "webm",
}

func isMediaFile(path string) bool {
	extDot := filepath.Ext(path)
	if len(extDot) < 2 {
		return false
	}
	ext := extDot[1:]
	for _, allowed := range(acceptedFileTypes) {
		if ext == allowed {
			return true
		}
	}
	return false
}

func scanMedia(ctx context.Context, server *Server, mediaPath string) <-chan error {
	errChan := make(chan error, 1)
	ncpu := runtime.NumCPU()
	workChan := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < ncpu; i++ {
		wg.Add(1)
		go processMedia(server, &wg, workChan, errChan)
	}

	go func() {
		log.Printf("beginning scan of path %s\n", mediaPath)

		err := filepath.WalkDir(mediaPath, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				errChan <- fmt.Errorf("scanMedia WalkDirFunc: %w", err)
				return nil
			}
			if d.IsDir() && strings.Contains(d.Name(), ".git") {
				fmt.Printf("[#%s]", d.Name())
				return filepath.SkipDir
			} else if d.IsDir() {
				fmt.Printf("[%s]", path)
				return nil
			} else if !isMediaFile(path) || !d.Type().IsRegular() {
				return nil
			}

			workChan <- path

			return nil
		})

		close(workChan)
		wg.Wait()

		fmt.Println()
		if err != nil {
			errChan <- fmt.Errorf("scanMedia: %w", err)
		} else {
			log.Println("finished scanning files")
		}

		close(errChan)
	}()

	return errChan
}

func processMedia(server *Server, wg *sync.WaitGroup, workChan <-chan string, errChan chan<- error) {
	for path := range(workChan) {
		fileData, err := ioutil.ReadFile(path)
		if err != nil {
			errChan <- fmt.Errorf("processMedia error reading file \"%s\": %w ", path, err)
			fmt.Printf("!")
			continue
		}
		sha1sum := sha1.Sum(fileData)
		sha1hex := fmt.Sprintf("%x", sha1sum)
		_, err = server.InsertMedia(path, sha1hex)
		if err != nil {
			errChan <- fmt.Errorf("processMedia failed to insert scanned media: %w", err)
			fmt.Printf("!")
		} else {
			fmt.Printf(".")
		}
	}
	wg.Done()
}
