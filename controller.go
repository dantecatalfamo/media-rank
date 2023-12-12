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
	tmpl, err := template.New("index").Parse(indexView)
	if err != nil {
		http.Error(w, "failed to parse template", 500)
		log.Printf("Controller.Index failed to parse index template: %s", err)
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
	if err := tmpl.Execute(w, tmplArgs); err != nil {
		http.Error(w, "failed to execute template", 500)
		log.Printf("Controller.Index failed to execute template: %s", err)
		return
	}
}

func (c *Controller) Media(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/media/"))
	if err != nil {
		http.Error(w, "invalid media id", 400)
		return
	}
	mediaInfo, err := c.s.GetMediaInfo(int64(id))
	if err != nil {
		log.Printf("Controller.Media failed to retrieve media (%d) from db: %s", id, err)
		http.Error(w, "invalid media id", 400)
		return
	}

	f, err := os.Open(mediaInfo.Path)
	if err != nil {
		log.Printf("Controller.Media failed to open media file at \"%s\": %s", mediaInfo.Path, err)
		http.Error(w, "failed to retrieve media", 500)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Printf("Controller.Media failed to stat file \"%s\": %s", mediaInfo.Path, err)
		http.Error(w, "failed to stat file", 500)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	// default response writer automatically detects mime types
	_, err = io.Copy(w, f)
	if err != nil {
		log.Printf("Controller.Media failed to write media file: %s", err)
		http.Error(w, "failed to write media file", 500)
		return
	}
}

func (c *Controller) Vote(w http.ResponseWriter, r *http.Request) {
	winner := r.FormValue("winner")
	loser := r.FormValue("loser")
	winnerId, err := strconv.Atoi(winner)
	if err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	loserId, err := strconv.Atoi(loser)
	if err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	log.Printf("winner: %s, loser: %s", winner, loser)
	if err := c.s.UpdateScores(int64(winnerId), int64(loserId)); err != nil {
		log.Printf("Controller.Vote failed to update scores. winner: %d, loser: %d", winnerId, loserId)
		http.Error(w, "error updating database", 500)
		return
	}
	http.Redirect(w, r, "/", 302)
}

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("list").Parse(listView)
	if err != nil {
		log.Printf("Controller.List failed to parse template: %s", err)
		http.Error(w, "internal error", 500)
		return
	}
	list, err := c.s.SortedList(true)
	if err != nil {
		log.Printf("Controller.List failed to get sorted list: %s", err)
		http.Error(w, "DB failure", 500)
		return
	}
	args := struct { List []MediaInfo }{ List: list }
	if err := tmpl.Execute(w, args); err != nil {
		http.Error(w, "failed to execute template", 500)
		log.Printf("Controller.List failed to execute template: %s", err)
		return
	}
}
