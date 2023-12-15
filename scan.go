package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
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

	go func() {
		log.Printf("beginning scan of path %s\n", mediaPath)

		err := filepath.WalkDir(mediaPath, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				log.Printf("scanMedia WalkDirFunc: %s", err)
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
			fileData, err := ioutil.ReadFile(path)
			if err != nil {
				log.Printf("error reading file: %s \"%s\"", err, path)
				return nil
			}
			sha1sum := sha1.Sum(fileData)
			sha1hex := fmt.Sprintf("%x", sha1sum)
			_, err = server.InsertMedia(path, sha1hex)
			if err != nil {
				return fmt.Errorf("failed to insert scanned media: %w", err)
			}
			fmt.Printf(".")

			return nil
		})
		fmt.Println()
		if err != nil {
			log.Printf("scanMedia: %s\n", err)
		} else {
			log.Println("finished scanning files")
		}
		errChan <- err
	}()

	return errChan
}
