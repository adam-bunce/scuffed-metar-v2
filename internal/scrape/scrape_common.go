package scrape

import (
	"fmt"
	"strings"
)

// TODO: need a "parallelize all of these, send results (all the same) to channel generic func"

type WeatherReport struct {
	Metar []string
	Taf   []string
}

// NOTE(adam); we can do a switch with this to generate urls for EACH site!
func NewUrlBuilder() *NavCanUrl {
	b := strings.Builder{}
	b.WriteString(NavCanBaseApiUrl)
	return &NavCanUrl{Builder: b, queryParams: make(map[string]string)}
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

func (n *NavCanUrl) Query(k, v string) *NavCanUrl {
	n.queryParams[k] = v
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
