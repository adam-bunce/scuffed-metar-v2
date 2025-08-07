package scrape

import (
	"fmt"
	"golang.org/x/net/html"
	"os"
	"strings"
	"testing"
)

func TestGetHighwaysWeatherReport(t *testing.T) {
	report, err := GetHighwaysWeatherReport("CZFD", "fonddulac")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(report)
}

func TestProcessHighwaysMetarResponse(t *testing.T) {
	documentRaw, err := os.ReadFile("testdata/happy_path/highways_cjy4.html")
	if err != nil {
		t.Fatal(err)
	}
	document, err := html.Parse(strings.NewReader(string(documentRaw)))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ProcessHighwaysMetarResponse(document)
	if err != nil {
		t.Fatal(err)
	}

	expectedMetarCount := 6
	expectedCamURLs := 3

	if len(result.Metar) != expectedMetarCount {
		t.Fatalf("Expected %d metars, got %d", expectedMetarCount, len(result.Metar))
	}
	if len(result.Cams) != expectedCamURLs {
		t.Fatalf("Expected %d cams, got %d", expectedCamURLs, len(result.Cams))
	}
}
