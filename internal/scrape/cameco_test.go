package scrape

import (
	"fmt"
	"testing"
)

// for _, airportCode := range []string{"CJW7", "CYKC", "CKQ8"} {
func TestGetCamecoWeatherReport(t *testing.T) {
	// takes 18 seconds btw
	report, err := GetCamecoWeatherReport("CJW7")
	if err != nil {
		return
	}
	fmt.Println(report)
}
