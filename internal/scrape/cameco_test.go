package scrape

import (
	"fmt"
	"testing"
)

func TestGetCamecoWeatherReport(t *testing.T) {
	// takes 18 seconds btw
	report, err := GetCamecoWeatherReport("CJW7")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(report)
}
