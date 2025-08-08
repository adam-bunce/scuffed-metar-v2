package api

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"scuffed-v2/internal/scrape"
	"slices"
)

// TODO: all metars in one place
// func HandleGetMetar()

var registry = []scrape.RequestCoordinator{
	{
		SupportedSites:   scrape.Navcansites,
		BatchFunc:        scrape.GetNavCanWeatherReports,
		SupportsBatching: true,
	},
	{
		SupportedSites:   slices.Collect(maps.Keys(scrape.SiteNamesMap)),
		PullFunc:         scrape.GetHighwaysWeatherReport,
		SupportsBatching: false,
	},
}

// GetMetar
func GetMetar(w http.ResponseWriter, req *http.Request) {
	// all metar data is cronned?
	// i think? might as well just cron it though
	// how do i want to handle caching different data from dfiferent
	// services and keeping it al in sync and lettin gusers specify endponits that also mihgt not exist?
	out, _ := scrape.DoTheThing(registry, []string{"CYXE", "CYYL", "CJY4"})
	for _, rec := range out {
		fmt.Println(rec)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
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
