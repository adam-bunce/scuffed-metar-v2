package scrape

import (
	"encoding/json"
	"fmt"
	"scuffed-v2/internal/util"
	"strconv"
	"strings"
	"time"
)

// NavCanadaResponse is a general structure returned from all NavCanada endpoints often
// with each Data Text field containing escaped json
type NavCanadaResponse struct {
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
		Position      struct {
			PointReference string `json:"pointReference"`
			RadialDistance int    `json:"radialDistance"`
		} `json:"position"`
	} `json:"data"`
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
	CloudsWeather           []GFAMetadata
	IcingTurbulenceFreezing []GFAMetadata
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
	StartValidity time.Time
	EndValidity   time.Time
	Id            string // the Id  the image used in creating the URL
}

// testString produces a string to use in testing
func (g *GFAMetadata) testString() string {
	return "[" + g.StartValidity.Format(NavCanadaTimeFormat) +
		" " + g.EndValidity.Format(NavCanadaTimeFormat) +
		" " + g.Id + "]"
}

const (
	NavCanBaseApiUrl = "https://plan.navcanada.ca/weather/api/alpha/?"

	CloudForecast      = "GFA/CLDWX/GFACN32/"
	TurbulenceForecast = "GFA/TURBC/GFACN32/"

	NavCanadaTimeFormat    = "2006-01-02T15:04:05"
	NavCanadaTimeFormatAlt = "2006-01-02T15:04:05+00:00"
)

// GetGFAImageIds sends a request to NavCanada's severs synchronously to get GFA (Graphic Area Forecast) data
func GetGFAImageIds() (GFA, error) {
	var body NavCanadaResponse

	err := util.RequestAndParse(NavCanBaseApiUrl+"site=CYXE&image=GFA/CLDWX&image=GFA/TURBC", &body)
	if err != nil {
		return GFA{}, err
	}

	return ProcessGFAResponse(body)

}

// ProcessGFAResponse extracts GFA data contained in gfaRes's NavCanadaResponse's Data.Text field
func ProcessGFAResponse(gfaRes NavCanadaResponse) (GFA, error) {
	var res GFA

	for _, datum := range gfaRes.Data {
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
