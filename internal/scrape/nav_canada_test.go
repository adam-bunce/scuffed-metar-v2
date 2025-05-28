package scrape

import (
	"scuffed-v2/internal/util"
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
		var result NavCanadaResponse
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

func TestProcessMETARResponse(t *testing.T) {

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
