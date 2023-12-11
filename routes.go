package main

import "net/http"

func SetupRoutes(s *Server) {
	controller := Controller{ s: s }
	http.HandleFunc("/", controller.Index)
	http.HandleFunc("/media/", controller.Media)
}
