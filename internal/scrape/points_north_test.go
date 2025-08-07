package scrape

import (
	"fmt"
	"testing"
)

func TestGetPointsNorthWeatherReport(t *testing.T) {
	report, err := GetPointsNorthWeatherReport("CYNL")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(report)
}
