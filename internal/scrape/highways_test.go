package scrape

import (
	"fmt"
	"testing"
)

func TestGetHighwaysWeatherReport(t *testing.T) {
	report, err := GetHighwaysWeatherReport("CJY4", "sandybay")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(report)
}
