package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Controller struct {
	s *Server
}

type IndexArgs struct {
	Media1 MediaInfo
	Media2 MediaInfo
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		log.Printf("failed to parse index template: %s", err)
		return
	}
	media1, media2, err := c.s.SelectMediaForComparison()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmplArgs := IndexArgs{
		Media1: media1,
		Media2: media2,
	}
	tmpl.Execute(w, tmplArgs)
}

const indexTemplate = `
<!DOCTYPE html>
<html>
<head>
<title>Media Rank</title>
<style>
  .container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-gap: 1em;
  }
  img {
    max-width: 100%;
  }
  .image {
    text-align: center;
  }
</style>
</head>
<body>
<div class="container">
  <div class="image">
    <img src="/media/{{.Media1.Id}}">
  </div>
  <div class="image">
    <img src="/media/{{.Media2.Id}}">
  </div>
  <button>Left image</button>
  <button>Right image</button>
</div>
{{.}}
</body>
</html>
`

func (c *Controller) Media(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/media/"))
	if err != nil {
		http.Error(w, "invalid media id", 400)
		return
	}
	mediaInfo, err := c.s.GetMediaInfo(int64(id))
	if err != nil {
		log.Printf("failed to retrieve media (%d) from db: %s", id, err)
		http.Error(w, "invalid media id", 400)
		return
	}

	f, err := os.Open(mediaInfo.Path)
	if err != nil {
		log.Printf("failed to open media file at \"%s\": %s", mediaInfo.Path, err)
		http.Error(w, "failed to retrieve media", 500)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Printf("failed to stat file \"%s\": %s", mediaInfo.Path, err)
		http.Error(w, "failed to stat file", 500)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	// default response writer automatically detects mime types
	_, err = io.Copy(w, f)
	if err != nil {
		log.Printf("failed to write media file: %s", err)
		http.Error(w, "failed to write media file", 500)
		return
	}
}
