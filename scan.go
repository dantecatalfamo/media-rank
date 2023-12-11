package main

import (
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

func scanMedia(server *Server, mediaPath string) error {
	log.Printf("beginning scan of path %s\n", mediaPath)
	filepath.WalkDir(mediaPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.Contains(d.Name(), ".git") {
			log.Printf("skipping %s\n", d.Name())
			return filepath.SkipDir
		} else if d.IsDir() {
			log.Printf("dir: %s\n", path)
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
		log.Printf("inserted %s into db\n", sha1hex)

		return nil
	})
	log.Println("finished scanning files")
	return nil
}
