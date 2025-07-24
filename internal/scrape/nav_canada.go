package scrape

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"scuffed-v2/internal/util"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	NavCanBaseApiUrl = "https://plan.navcanada.ca/weather/api/alpha/?"

	CloudForecast      = "GFA/CLDWX/GFACN32/"
	TurbulenceForecast = "GFA/TURBC/GFACN32/"

	NavCanadaTimeFormat    = "2006-01-02T15:04:05"
	NavCanadaTimeFormatAlt = "2006-01-02T15:04:05+00:00"
)

// NavCanadaResponse is a general structure returned from all NavCanada endpoints often
// with each Data Text field containing escaped json
type NavCanadaResponse[PositionType any] struct {
	Meta struct {
		Now   string `json:"now"`
		Count struct {
			Metar int `json:"metar"`
			Taf   int `json:"taf"`
		} `json:"count"`
		Messages []interface{} `json:"messages"`
	} `json:"meta"`
	Data []struct {
		Type          string `json:"type"`
		Pk            string `json:"pk"`
		Location      string `json:"location"`
		StartValidity string `json:"startValidity"`
		EndValidity   string `json:"endValidity"`
		Text          string `json:"text"` // escaped JSON to be further parsed
		HasError      bool   `json:"hasError"`
		// Positions can be either a list of Position, or a singular Position if only one site is requested
		Positions PositionType `json:"position"`
	} `json:"data"`
}

type Position struct {
	PointReference any `json:"pointReference"` // "CYXE", "0.9646", "0"
	RadialDistance any `json:"radialDistance"`
}

// GFAText represents the Text section of a NavCanadaResponse GFA query
type GFAText struct {
	Product      string `json:"product"`
	SubProduct   string `json:"sub_product"`
	Geography    string `json:"geography"`
	SubGeography string `json:"sub_geography"`
	FrameLists   []struct {
		Id     int    `json:"id"`
		Sv     string `json:"sv"`
		Ev     string `json:"ev"`
		Frames []struct {
			Id            int    `json:"id"`
			StartValidity string `json:"sv"`
			EndValidity   string `json:"ev"`
			Images        []struct {
				Id      int    `json:"id"`
				Created string `json:"created"`
			} `json:"images"`
		} `json:"frames"`
	} `json:"frame_lists"`
}

// GFA is the desired data extracted from a NavCanadaResponse
type GFA struct {
	CloudsWeather           []GFAMetadata `json:"clouds_weather"`
	IcingTurbulenceFreezing []GFAMetadata `json:"icing_turbulence_freezing"`
}

// testString produces a string to use in testing
func (g *GFA) testString() string {
	builder := strings.Builder{}
	for _, val := range g.CloudsWeather {
		builder.WriteString(val.testString())
	}

	for _, val := range g.IcingTurbulenceFreezing {
		builder.WriteString(val.testString())
	}

	return builder.String()
}

// A GFAMetadata contains the minimal information needed to display and select GFA
type GFAMetadata struct {
	StartValidity time.Time `json:"start_validity"`
	EndValidity   time.Time `json:"end_validity"`
	Id            string    `json:"id"` // the Id  the image used in creating the URL (TODO: maybe do server-side?
}

// testString produces a string to use in testing
func (g *GFAMetadata) testString() string {
	return "[" + g.StartValidity.Format(NavCanadaTimeFormat) +
		" " + g.EndValidity.Format(NavCanadaTimeFormat) +
		" " + g.Id + "]"
}

// GetGFAImageIds sends a request to NavCanada's severs synchronously to get GFA (Graphic Area Forecast) data
func GetGFAImageIds() (GFA, error) {
	var body NavCanadaResponse[Position]

	url := NewUrlBuilder().
		Sites("CYXE").
		Images(GfaTurbulence, GfaClouds).
		Build()

	err := util.GetAndParse(url, &body)
	if err != nil {
		return GFA{}, err
	}

	return ProcessGFAResponse(body)

}

// ProcessGFAResponse extracts GFA data contained in gfaRes's NavCanadaResponse's Data.Text field
func ProcessGFAResponse(gr NavCanadaResponse[Position]) (GFA, error) {
	var res GFA

	for _, datum := range gr.Data {
		gfaMeta, err := ExtractGFAMeta(datum.Text)
		if err != nil {
			return GFA{}, err
		}

		switch datum.Location {
		case CloudForecast:
			res.CloudsWeather = append(res.CloudsWeather, gfaMeta...)
		case TurbulenceForecast:
			res.IcingTurbulenceFreezing = append(res.IcingTurbulenceFreezing, gfaMeta...)
		default:
			return GFA{}, fmt.Errorf("unknown location: %s", datum.Location)
		}
	}
	return res, nil
}

// ExtractGFAMeta extracts each frames data from the last FramesList from GFAText into GFAMetadata
func ExtractGFAMeta(text string) ([]GFAMetadata, error) {
	var gfaText GFAText
	var records []GFAMetadata
	err := json.NewDecoder(strings.NewReader(text)).Decode(&gfaText)
	if err != nil {
		return nil, err
	}

	hasFrames := len(gfaText.FrameLists) >= 0
	if !hasFrames {
		return nil, fmt.Errorf("no frames found")
	}

	lastFrameIdx := len(gfaText.FrameLists) - 1
	lastFrames := gfaText.FrameLists[lastFrameIdx].Frames

	for _, frame := range lastFrames {
		if len(frame.Images) > 0 {
			meta := GFAMetadata{}

			meta.Id = strconv.Itoa(frame.Images[len(frame.Images)-1].Id)

			meta.EndValidity, err = time.Parse(NavCanadaTimeFormat, frame.EndValidity)
			if err != nil {
				return nil, err
			}

			meta.StartValidity, err = time.Parse(NavCanadaTimeFormat, frame.StartValidity)
			if err != nil {
				return nil, err
			}

			records = append(records, meta)
		}
	}

	return records, nil
}

// GetNavCanWeatherReports returns the metar and taf readouts for the specified sites
func GetNavCanWeatherReports(sites ...string) ([]*WeatherReport, error) {
	var body NavCanadaResponse[any]

	url := NewUrlBuilder().
		Sites(sites...).
		MetarChoice(3).
		Alpha(Metar, Taf).
		Build()

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	err = util.RequestAndParse(request, &body)
	if err != nil {
		return nil, err
	}

	reports, err := ProcessMETARResponse(body)
	return slices.Collect(maps.Values(reports)), err
}

// ProcessMETARResponse processes a METAR records for single or multiple unique sites
func ProcessMETARResponse(mr NavCanadaResponse[any]) (map[string]*WeatherReport, error) {
	res := make(map[string]*WeatherReport)

	for _, datum := range mr.Data {
		airportCode := datum.Location
		if res[airportCode] == nil {
			res[airportCode] = &WeatherReport{Airport: airportCode}
		}
		switch Alpha(datum.Type) {
		case Metar:
			res[airportCode].Metar = append(res[airportCode].Metar, datum.Text)
		case Taf:
			res[airportCode].Taf = append(res[airportCode].Taf, datum.Text)
		default:
			slog.Info("Skipping unknown metar", slog.String("type", datum.Type))
		}
	}

	return res, nil
}

type ImageType string

const (
	GfaClouds     ImageType = "GFA/CLDWX"
	GfaTurbulence ImageType = "GFA/TURBC"
)

type Alpha string

const (
	Airmet    Alpha = "airmet"
	Sigmet    Alpha = "sigmet"
	Metar     Alpha = "metar"
	Taf       Alpha = "taf"
	Upperwind Alpha = "upperwind"
)

type NavCanUrl struct {
	Builder     strings.Builder
	sites       []string
	metarChoice int
	alpha       []Alpha
	queryParams map[string]string
	imageTypes  []ImageType
	radius      int
}

type AirportWinds struct {
	AirportCode string `json:"airport_code"`

	High []Wind `json:"high_winds"`
	Low  []Wind `json:"low_winds"`
}

type Wind struct {
	Data []ElevationValues `json:"elevation_values"`

	BasedOn     time.Time `json:"based_on"`
	Valid       time.Time `json:"valid"`
	ForUseStart time.Time `json:"for_use_start"`
	ForUseEnd   time.Time `json:"for_use_end"`
}
type ElevationValues struct {
	Elevation *float64 `json:"elevation"`
	// Values is a *float64 because it can be empty
	Values []*float64 `json:"values"`
}

func GetWinds(sites ...string) ([]AirportWinds, error) {
	var body NavCanadaResponse[any]

	url := NewUrlBuilder().
		Sites(sites...).
		Alpha(Upperwind).
		Query("upperwind_choice", "both").
		Build()

	fmt.Println("url", url)

	err := util.GetAndParse(url, &body)
	if err != nil {
		return nil, err
	}

	fmt.Println("err?", err)

	return ProcessWindsResponse(body)
}

type WindsText struct {
	Numbers []int
	Strings []string
	Arrays  [][]*float64
	Times   []time.Time
}

// UnmarshalJSON overrides the default json parse function based on struct annotations as WindsText is
// []any and cannot be accurately described via struct tags
func (w *WindsText) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty, nothing to unmarshal")
	}

	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for i, entry := range raw {
		// Returned data always has "null" for in 7/8/9/10 positions
		if slices.Contains([]int{7, 8, 9, 10}, i) {
			continue
		}

		switch entry.(type) {
		case string:
			entryStr, ok := entry.(string)
			if !ok {
				return fmt.Errorf("entry is not a string")
			}
			timestamp, err := time.Parse(NavCanadaTimeFormatAlt, entryStr)
			if err == nil {
				w.Times = append(w.Times, timestamp.UTC())
			} else {
				w.Strings = append(w.Strings, entryStr)
			}
		case float64:
			entryFloat, ok := entry.(float64)
			if !ok {
				return fmt.Errorf("cannot unmarshal float64")
			}
			w.Numbers = append(w.Numbers, int(entryFloat))
		case int:
			entryInt, ok := entry.(int)
			if !ok {
				return fmt.Errorf("cannot unmarshal int")
			}
			w.Numbers = append(w.Numbers, entryInt)
		case []interface{}:
			// winds data in [][]int
			windsArr, ok := entry.([]interface{})
			if !ok {
				return fmt.Errorf("windsArray is not []interface{}")
			}
			for _, windsArray := range windsArr {
				var parsedWindsValues []*float64
				windsSubArray, ok := windsArray.([]interface{})
				if !ok {
					return fmt.Errorf("windsSubArray is not []interface{}")
				}
				for _, windArray := range windsSubArray {
					windsSubArrayFloat, ok := windArray.(float64)
					if !ok {
						// set as null, probably is
						parsedWindsValues = append(parsedWindsValues, nil)
						continue
					}
					parsedWindsValues = append(parsedWindsValues, &windsSubArrayFloat)
				}
				w.Arrays = append(w.Arrays, parsedWindsValues)
			}
		default:
			fmt.Printf("unknown type %T at %d\n", entry, i)
		}
	}

	return nil
}

const (
	expectedWindsCount = 5 // Elevation, <3x values>, 0
	elevationIndex     = 0
	lowThreshold       = 18_000.0
)

func ProcessWindsResponse(wr NavCanadaResponse[any]) ([]AirportWinds, error) {
	airportWinds := make(map[string]AirportWinds)

	// each wind record maps a wind state (full set of higher or lower and an associated timestamp) to an airport
	for _, windRecord := range wr.Data {
		currentAirport := windRecord.Location
		// create if doesnt exist
		_, ok := airportWinds[currentAirport]
		if !ok {
			airportWinds[currentAirport] = AirportWinds{AirportCode: currentAirport}
		}

		var wt WindsText
		err := json.NewDecoder(strings.NewReader(windRecord.Text)).Decode(&wt)
		if err != nil {
			return nil, err
		}

		var (
			highWinds Wind
			lowWinds  Wind
		)

		//  first value is elevation, the rest are the speeds
		for _, wind := range wt.Arrays {
			if len(wind) != expectedWindsCount {
				slog.Info("Expected winds count doesn't match actual",
					slog.Int("expected", expectedWindsCount),
					slog.Int("actual", len(wind)),
					slog.Any("arr", wind),
				)
				continue
			}

			// add elevation values based on height
			ev := ElevationValues{
				Elevation: wind[elevationIndex],
			}
			for _, windValue := range wind[elevationIndex+1:] {
				ev.Values = append(ev.Values, windValue)
			}

			if *ev.Elevation <= lowThreshold {
				lowWinds.Data = append(lowWinds.Data, ev)
			} else {
				highWinds.Data = append(highWinds.Data, ev)
			}
		}

		const (
			UnknownIndex     = 0
			BasedOnIndex     = 1
			ValidIndex       = 2
			ForUseStartIndex = 3
			ForUseEndIndex   = 4
		)

		if len(wt.Times) != 5 {
			slog.Info("Expected 5 times for winds", slog.Int("actual", len(wt.Times)))
			continue
		}

		// We're only operating one type of wind per-loop one, but set both
		lowWinds.BasedOn, highWinds.BasedOn = wt.Times[BasedOnIndex], wt.Times[BasedOnIndex]
		lowWinds.Valid, highWinds.Valid = wt.Times[ValidIndex], wt.Times[ValidIndex]
		lowWinds.ForUseStart, highWinds.ForUseStart = wt.Times[ForUseStartIndex], wt.Times[ForUseStartIndex]
		lowWinds.ForUseEnd, highWinds.ForUseEnd = wt.Times[ForUseEndIndex], wt.Times[ForUseEndIndex]

		// ignore nil's
		if highWinds.Data == nil {
			airportWinds[currentAirport] = AirportWinds{
				AirportCode: airportWinds[currentAirport].AirportCode,
				High:        airportWinds[currentAirport].High,
				Low:         append(airportWinds[currentAirport].Low, lowWinds),
			}
		} else {
			airportWinds[currentAirport] = AirportWinds{
				AirportCode: airportWinds[currentAirport].AirportCode,
				High:        append(airportWinds[currentAirport].High, highWinds),
				Low:         airportWinds[currentAirport].Low,
			}
		}
	}

	return slices.Collect(maps.Values(airportWinds)), nil
}
