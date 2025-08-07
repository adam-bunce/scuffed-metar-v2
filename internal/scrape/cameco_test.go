package scrape

import (
	"fmt"
	"testing"
)

func TestGetCamecoWeatherReport(t *testing.T) {
	// takes 18 seconds btw
	// ?no it doesnt?
	report, err := GetCamecoWeatherReport("CJW7")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(report)
}
