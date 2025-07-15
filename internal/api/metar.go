package api

import (
	"encoding/json"
	"net/http"
	"scuffed-v2/internal/scrape"
)

// TODO: all metars in one place
// func HandleGetMetar()

func GetNavCanMetar(w http.ResponseWriter, req *http.Request) {
	// TODO: need to add presentation logic for consistent sort ordering
	// all metar data is cronned?
	// i think? might as well just cron it though
	data, err := scrape.GetNavCanWeatherReports("CYXE") //, "CYSF")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(data)
}

func GetGFA(w http.ResponseWriter, req *http.Request) {
	data, err := scrape.GetGFAImageIds()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(data)
}

func GetWinds(w http.ResponseWriter, req *http.Request) {
	data, err := scrape.GetWinds("CYXE")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(data)
}
