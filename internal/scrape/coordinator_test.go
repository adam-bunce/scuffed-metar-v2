package scrape

import (
	"fmt"
	"maps"
	"slices"
	"testing"
)

var registry = []RequestCoordinator{
	{
		SupportedSites:   Navcansites,
		BatchFunc:        GetNavCanWeatherReports,
		SupportsBatching: true,
	},
	{
		SupportedSites:   slices.Collect(maps.Keys(SiteNamesMap)),
		PullFunc:         GetHighwaysWeatherReport,
		SupportsBatching: false,
	},
}

func TestDoTheThing(t *testing.T) {
	fmt.Println(registry[1].SupportedSites)
	out, err := DoTheThing(registry, []string{"CYXE", "CYYL", "CJY4"})
	if err != nil {
		t.Fatal(err)
	}

	for _, rec := range out {
		fmt.Println(rec)
	}
}
