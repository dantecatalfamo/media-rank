package main

import "net/http"

func SetupRoutes(s *Server) {
	controller := Controller{ s: s }
	http.HandleFunc("/", controller.Index)
	http.HandleFunc("/media/", controller.Media)
	http.HandleFunc("/vote", controller.Vote)
	http.HandleFunc("/list", controller.List)
	http.HandleFunc("/history", controller.History)
}
