package api

import (
	"encoding/json"
	"net/http"
	"scuffed-v2/internal/scrape"
)

// TODO: all metars in one place
// func HandleGetMetar()

// GetMetar
func GetMetar(w http.ResponseWriter, req *http.Request) {
	// all metar data is cronned?
	// i think? might as well just cron it though
	// how do i want to handle caching different data from dfiferent
	// services and keeping it al in sync and lettin gusers specify endponits that also mihgt not exist?
	data, err := scrape.GetNavCanWeatherReports("CYXE", "CYSF")
	// TODO: need to add presentation logic for consistent sort ordering
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
