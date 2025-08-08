package scrape

import (
	"fmt"
	"golang.org/x/net/html"
	"log/slog"
	"scuffed-v2/internal/util"
	"strings"
)

var SiteNamesMap = map[string]string{
	"CYBE": "uranium",
	"CZFD": "fonddulac",
	"CYSF": "stonyrapids",
	"CZWL": "wollaston",

	"CJL4": "laloche",
	"CYVT": "buffalonarrows",
	"CKB2": "patuanak",
	"CJF3": "ilealacrosse",

	"CZPO": "pinehouse",
	"CJY4": "sandybay",
	"CJW4": "pelican",
	"CJT4": "cumberlandhouse",

	"CYLJ": "meadowlake",
	"CYHB": "hudsonbay",
}

func GetHighwaysWeatherReport(site string) (*WeatherReport, error) {
	siteName := SiteNamesMap[site]
	slog.Info("highways", slog.String("siteName", siteName), slog.String("site", site))
	var body string
	url := fmt.Sprintf("http://highways.glmobile.com/%s", siteName)
	err := util.GetAndParseString(url, &body)
	if err != nil {
		return nil, err
	}
	document, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	return ProcessHighwaysMetarResponse(document, url, site)
}

func ProcessHighwaysMetarResponse(document *html.Node, url, site string) (*WeatherReport, error) {
	res := &WeatherReport{
		Airport: site,
		Cams:    ExtractCamUrls(document, url),
		Metar:   ExtractMetarReadOuts(document),
	}
	return res, nil
}

// TODO: are there ever TAF's? i feel like no - not sure though
func ExtractMetarReadOuts(root *html.Node) []string {
	var res []string

	var f func(n *html.Node)
	f = func(n *html.Node) {
		// base case
		if n == nil {
			return
		}

		isBoldNodeWithChild := false
		isBoldNodeWithChild = n.Type == html.ElementNode && n.Data == "b" && n.FirstChild != nil

		if isBoldNodeWithChild {
			childText := n.FirstChild.Data
			if strings.Contains(childText, "METAR") || strings.Contains(childText, "SPECI") || strings.Contains(childText, "LWIS") {
				res = append(res, n.FirstChild.Data)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	return res
}

func ExtractCamUrls(root *html.Node, domain string) []string {
	var res []string

	// visit all nodes and if it's an image tag add its src attribute
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					res = append(res, domain+"/"+a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	return res
}
