package scrape

import (
	"fmt"
	"log/slog"
	"regexp"
	"scuffed-v2/internal/util"
)

var pointsNorthRegex = regexp.MustCompile(`(?i)<TD COLSPAN="3">(.*?)</TD>`)

func GetPointsNorthWeatherReport(site string) (*WeatherReport, error) {
	var data string

	err := util.GetAndParseString(fmt.Sprintf("https://www.pointsnorthgroup.ca/weather/%s_metar.html", site), &data)
	if err != nil {
		return nil, err
	}

	return ProcessPointsNorthMetarResponse(data, site)
}

func ProcessPointsNorthMetarResponse(mr string, site string) (*WeatherReport, error) {
	res := WeatherReport{
		Airport: site,
	}

	matches := pointsNorthRegex.FindAllStringSubmatch(mr, -1)
	if len(matches) < 0 {
		slog.Error("expected more matches in points north request", slog.Int("actual", len(matches)))
		return nil, fmt.Errorf("expected more matches in points north request")
	}

	for _, match := range matches {
		res.Metar = append(res.Metar, match[1])
	}

	return &res, nil
}
