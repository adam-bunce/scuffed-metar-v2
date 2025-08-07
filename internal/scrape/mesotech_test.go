package scrape

import (
	"fmt"
	"testing"
)

func TestGetMesotechWeatherReport(t *testing.T) {
	report, err := GetMesotechWeatherReport("CET2")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(report)
}
