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

type MediaInfo struct {
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
	Score int   `json:"score"`
}

var acceptedFileTypes = []string{
	"jpg",
	"png",
	"gif",
	"mp4",
	"webm",
}

const insertScan = `
INSERT INTO media(path, sha1sum, score) VALUES (?, ?, 1500)
  ON CONFLICT(sha1sum) DO UPDATE SET path = ?
`

func isMediaFile(path string) bool {
	ext := filepath.Ext(path)[1:]
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
		} else if !isMediaFile(path) {
			return nil
		}
		fileData, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("error reading file: %s \"%s\"", err, path)
			return nil
		}
		sha1sum := sha1.Sum(fileData)
		sha1hex := fmt.Sprintf("%x", sha1sum)
		server.db.Exec(insertScan, path, sha1hex, path)
		log.Printf("inserted %s into db\n", sha1hex)

		return nil
	})
	return nil
}
