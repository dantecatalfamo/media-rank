package main

import (
	"fmt"
	"net/http"
)

type Controller struct {
	s *Server
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	media1, media2, err := c.s.SelectMediaForComparison()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "media1: %+v\nmedia2: %+v", media1, media2)
}
