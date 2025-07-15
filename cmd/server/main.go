package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"scuffed-v2/internal/api"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/metar", api.GetNavCanMetar)
	r.HandleFunc("/gfa", api.GetGFA)
	r.HandleFunc("/winds", api.GetWinds)

	log.Fatal(http.ListenAndServe(":8080", r))
}
