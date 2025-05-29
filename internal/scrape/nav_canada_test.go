package scrape

import (
	"maps"
	"scuffed-v2/internal/util"
	"slices"
	"testing"
)

func TestProcessGFAResponse(t *testing.T) {
	cases := []struct {
		testFilePath string
		expected     string
	}{
		{
			"testdata/happy_path/gfa_response.json",
			"[2025-05-18T00:00:00 2025-05-18T06:00:00 56731137]" +
				"[2025-05-18T06:00:00 2025-05-18T12:00:00 56731152]" +
				"[2025-05-18T12:00:00 2025-05-18T18:00:00 56731163]" +
				"[2025-05-18T00:00:00 2025-05-18T06:00:00 56731145]" +
				"[2025-05-18T06:00:00 2025-05-18T12:00:00 56731150]" +
				"[2025-05-18T12:00:00 2025-05-18T18:00:00 56731173]",
		},
	}

	for _, tc := range cases {
		var result NavCanadaResponse[Position]
		err := util.ReadFileToStruct(tc.testFilePath, &result)
		if err != nil {
			t.Fatalf("Could not access test data, %s", err)
		}

		actual, err := ProcessGFAResponse(result)
		if err != nil {
			t.Fatalf("Unexpected error processing GFA response %s", err)
		}

		if actual.testString() != tc.expected {
			t.Fatalf("\nExpected: %v\nActual:   %v", tc.expected, actual.testString())
		}
	}
}

func TestNavCanUrl_GetUrl(t *testing.T) {
	actual := NewUrlBuilder().
		Sites("CYXE", "CYSF").
		MetarChoice(3).
		Alpha(Metar, Taf).
		Build()

	expected := "https://plan.navcanada.ca/weather/api/alpha/?&site=CYXE&site=CYSF&metar_choice=3&alpha=metar&alpha=taf&radius=0"

	if expected != actual {
		t.Fatalf("expected %q got %q", expected, actual)
	}
}

func TestGetWeatherReports(t *testing.T) {
	expectedSites := []string{"CYXE", "CYSF"}

	sites, err := GetWeatherReports(expectedSites...)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(slices.Collect(maps.Keys(sites)), expectedSites) {
		t.Fatalf("expected map with 2 keys got %d keys", len(slices.Collect(maps.Keys(sites))))
	}
}

func TestGetGFAImageIds(t *testing.T) {
	gfa, err := GetGFAImageIds()
	if err != nil {
		t.Fatal(err)
	}

	if len(gfa.IcingTurbulenceFreezing) != 3 || len(gfa.CloudsWeather) != 3 {
		t.Fatalf("Expected 3 images for both IcingTurbulenceFreezing(%d) and CloudsWeather(%d)", len(gfa.IcingTurbulenceFreezing), len(gfa.CloudsWeather))
	}

}
