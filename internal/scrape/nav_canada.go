package scrape

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"scuffed-v2/internal/util"
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
type NavCanadaResponse[PosType any] struct {
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
		Positions PosType `json:"position"`
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

// GetGFAImageIds sends a request to NavCanada's severs synchronously to get GFA (Graphic Area Forecast) data
func GetGFAImageIds() (GFA, error) {
	var body NavCanadaResponse[Position]

	url := NewUrlBuilder().
		Sites("CYXE").
		Images(GfaTurbulence, GfaClouds).
		Build()

	err := util.RequestAndParse(url, &body)
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

type WeatherReport struct {
	Metar []string
	Taf   []string
}

func GetWeatherReports(sites ...string) (map[string]*WeatherReport, error) {
	var body NavCanadaResponse[[]Position]

	url := NewUrlBuilder().
		Sites(sites...).
		MetarChoice(3).
		Alpha(Metar, Taf).
		Build()

	err := util.RequestAndParse(url, &body)
	if err != nil {
		return nil, err
	}

	return ProcessMETARResponse(body)

}

// ProcessMETARResponse processes a METAR records for single or multiple unique sites
func ProcessMETARResponse(mr NavCanadaResponse[[]Position]) (map[string]*WeatherReport, error) {
	res := make(map[string]*WeatherReport)

	for _, datum := range mr.Data {
		airportCode := datum.Location
		if res[airportCode] == nil {
			res[airportCode] = &WeatherReport{}
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
	Airmet Alpha = "airmet"
	Sigmet Alpha = "sigmet"
	Metar  Alpha = "metar"
	Taf    Alpha = "taf"
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

func NewUrlBuilder() *NavCanUrl {
	b := strings.Builder{}
	b.WriteString(NavCanBaseApiUrl)
	return &NavCanUrl{Builder: b}
}

func (n *NavCanUrl) Sites(sites ...string) *NavCanUrl {
	n.sites = append(n.sites, sites...)
	return n
}

func (n *NavCanUrl) MetarChoice(choice int) *NavCanUrl {
	n.metarChoice = choice
	return n
}

func (n *NavCanUrl) Alpha(alpha ...Alpha) *NavCanUrl {
	n.alpha = append(n.alpha, alpha...)
	return n
}

func (n *NavCanUrl) Images(imageTypes ...ImageType) *NavCanUrl {
	n.imageTypes = append(n.imageTypes, imageTypes...)
	return n
}

func (n *NavCanUrl) Radius(radius int) *NavCanUrl {
	n.radius = radius
	return n
}

func (n *NavCanUrl) Query(queryParams map[string]string) *NavCanUrl {
	for k, v := range queryParams {
		n.queryParams[k] = v
	}
	return n
}

func (n *NavCanUrl) Build() string {
	builder := strings.Builder{}
	builder.WriteString(NavCanBaseApiUrl)

	for _, site := range n.sites {
		builder.WriteString(fmt.Sprintf("&site=%s", strings.ToUpper(site)))
	}

	builder.WriteString(fmt.Sprintf("&metar_choice=%d", n.metarChoice))

	for _, alpha := range n.alpha {
		builder.WriteString(fmt.Sprintf("&alpha=%s", alpha))
	}

	for _, img := range n.imageTypes {
		builder.WriteString(fmt.Sprintf("&image=%s", img))
	}

	for key, value := range n.queryParams {
		builder.WriteString(fmt.Sprintf("&%s=%s", key, value))
	}

	builder.WriteString(fmt.Sprintf("&radius=%d", n.radius))

	return builder.String()
}
