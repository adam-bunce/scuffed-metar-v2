package scrape

import (
	"fmt"
	"log/slog"
	"slices"
)

var Navcansites = []string{
	"CYXE",
	"CYVT",
	"CYLJ",
	"CYSF",
	"CYVC",
	"CYKJ",
	"CYPA",
	"CYFO",
	"CYQW",
	"CYQR",
	"CYMM",
	"CYSM",
	"CYPY",
	"CYQD",
	"CYLL",
	"CYYN",
	"CYXH",
	"CYTH",
	"CYQV",
	"CYOD",
	"CYYL",
}

type RequestCoordinator struct {
	SupportedSites   []string
	PullFunc         func(string) (*WeatherReport, error)
	BatchFunc        func([]string) ([]*WeatherReport, error)
	SupportsBatching bool
}

func DoTheThing(c []RequestCoordinator, s []string) ([]*WeatherReport, error) {
	remaining := slices.Clone(s) // work with copy
	var res []*WeatherReport
	// batch what we can
	for _, coordinator := range c {
		if !coordinator.SupportsBatching {
			continue
		}

		var batchable []string
		for _, site := range s {
			if slices.Contains(coordinator.SupportedSites, site) {
				batchable = append(batchable, site)
				slices.DeleteFunc(remaining, func(s string) bool {
					return s == site
				})
			}
		}

		if coordinator.SupportsBatching {
			fmt.Println("batching", batchable)
			results, err := coordinator.BatchFunc(batchable)
			if err != nil {
				slog.Error("Unable to handle batch", slog.String("err", err.Error()), slog.Any("batchable", batchable))
				panic(err)
			}
			res = append(res, results...)
		}
	}

	// do the rest as singles
	for _, site := range remaining {
		for _, coordinator := range c {
			if slices.Contains(coordinator.SupportedSites, site) {
				out, err := coordinator.PullFunc(site)
				if err != nil {
					panic(err)
				}
				res = append(res, out)
			}
		}
	}

	return res, nil
}
