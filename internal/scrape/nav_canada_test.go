package scrape

import (
	"encoding/json"
	"reflect"
	"scuffed-v2/internal/util"
	"strings"
	"testing"
	"time"
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

func TestNavCanUrl_BuildWinds(t *testing.T) {
	actual := NewUrlBuilder().
		Sites("CYQR", "CYVC", "CYXE", "CYYL").
		Alpha(Upperwind).
		Query("upperwind_choice", "both").
		Build()

	expected := "https://plan.navcanada.ca/weather/api/alpha/?&site=CYQR&site=CYVC&site=CYXE&site=CYYL&metar_choice=0&alpha=upperwind&upperwind_choice=both&radius=0"
	if expected != actual {
		t.Fatalf("expected %q got %q", expected, actual)
	}
}

func TestGetWeatherReports(t *testing.T) {
	expectedSites := []string{"CYXE", "CYSF"}

	sites, err := GetNavCanWeatherReports(expectedSites)
	if err != nil {
		t.Fatal(err)
	}

	if len(sites) != len(expectedSites) {
		t.Fatalf("expected array with 2 entries got %d entries", len(sites))
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

func TestGetWinds(t *testing.T) {
	// want to test the different formats we get for location during parsing (can be array or single item)
	cases := []struct {
		sites []string
	}{
		{sites: []string{"CYXE", "CYYL"}},
		{sites: []string{"CYQR"}},
		{sites: []string{""}},
	}

	for _, tc := range cases {
		_, err := GetWinds(tc.sites...)
		if err != nil {
			t.Fatalf("Should be able to get winds without error: %q", err)
		}
	}
}

// TestParseWindsText tests the custom json decoder for winds text, which takes an array of mixed data types and partitions it
func TestParseWindsText(t *testing.T) {
	input := "[\"FBCN35\", \"KWNO\", " +
		"\"2025-06-08T13:57:00+00:00\", \"2025-06-08T12:00:00+00:00\", \"2025-06-09T12:00:00+00:00\", \"2025-06-09T06:00:00+00:00\", \"2025-06-09T18:00:00+00:00\", " +
		"null, null, null, null, " +
		"[[45000,330,34,-59,0],[53000,330,20,-58,0],[39000,310,58,-57,0],[34000,320,54,-48,0],[30000,310,51,-39,0],[24000,310,51,-24,0]]]"

	f2p := func(f float64) *float64 { return &f }
	expected :=
		WindsText{
			Numbers: []int{},
			Strings: []string{"FBCN35", "KWNO"},
			Arrays: [][]*float64{
				{f2p(45000), f2p(330), f2p(34), f2p(-59), f2p(0)},
				{f2p(53000), f2p(330), f2p(20), f2p(-58), f2p(0)},
				{f2p(39000), f2p(310), f2p(58), f2p(-57), f2p(0)},
				{f2p(34000), f2p(320), f2p(54), f2p(-48), f2p(0)},
				{f2p(30000), f2p(310), f2p(51), f2p(-39), f2p(0)},
				{f2p(24000), f2p(310), f2p(51), f2p(-24), f2p(0)}},
			Times: []time.Time{},
		}

	var wt WindsText
	err := json.NewDecoder(strings.NewReader(input)).Decode(&wt)
	if err != nil {
		t.Fatal("UnmarshallJSON unable to parse input", err)
	}

	// check arrays
	if len(wt.Arrays) != len(expected.Arrays) {
		t.Errorf("expected %d arrays, got %d", len(expected.Arrays), len(wt.Arrays))
	} else {
		for i := range len(wt.Arrays) {
			if !reflect.DeepEqual(wt.Arrays[i], expected.Arrays[i]) {
				t.Errorf("expected array %+v, got %+v", expected.Arrays[i], wt.Arrays[i])
			}
		}
	}

	// check strings
	if len(wt.Strings) != len(expected.Strings) {
		t.Errorf("expected %d strings, got %d", len(expected.Strings), len(wt.Strings))
		t.Errorf("expected: %v\n", expected.Strings)
		t.Errorf("got: %v\n", wt.Strings)
	} else {
		for i := range len(wt.Strings) {
			if wt.Strings[i] != expected.Strings[i] {
				t.Errorf("expected  string %q, got %q", expected.Strings[i], wt.Strings[i])
			}
		}
	}
}
