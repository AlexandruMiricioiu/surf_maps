package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const SURF_MAPS_URL = "https://surfheaven.eu/player/43223876"
const SURFHEAVEN_URL = "https://surfheaven.eu"

type surfMap struct {
	url         string
	completions int
	timesPlayed int
	tier        int
	kind        string
	bonuses     int
	checkpoints int
}

func getSurfMapUrls() ([]string, error) {
	res, err := http.Get(SURF_MAPS_URL)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("SurfHeaven request returned error status: (code %d) (status %s)", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing HTML for SurfHeaven maps")
	}

	mapUrls := make([]string, 0)
	doc.Find(".table-maps .table tbody tr td a").Each(func(i int, s *goquery.Selection) {
		mapHref, ok := s.Attr("href")
		if ok {
			mapUrls = append(mapUrls, mapHref)
		}
	})

	sort.Strings(mapUrls)

	return mapUrls[:10], nil
}

func getSurfMap(mapUrl string) (surfMap, error) {
	res, err := http.Get(SURFHEAVEN_URL + mapUrl)
	if err != nil {
		return surfMap{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return surfMap{}, fmt.Errorf("SurfHeaven request returned error status: (code %d) (status %s)", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return surfMap{}, fmt.Errorf("Error parsing HTML for SurfHeaven map url")
	}

	elements := doc.Find(".table.table-responsive.nodatatable tbody tr td")

	var m surfMap

	m.url = mapUrl
	fmt.Sscanf(strings.TrimSpace(elements.Eq(0).Text()), "%d Completions", &m.completions)
	fmt.Sscanf(strings.TrimSpace(elements.Eq(1).Text()), "%d Times Played", &m.timesPlayed)
	fmt.Sscanf(strings.TrimSpace(elements.Eq(2).Text()), "%d Tier", &m.tier)
	m.kind = strings.TrimSpace(elements.Eq(3).Text())
	fmt.Sscanf(strings.TrimSpace(elements.Eq(4).Text()), "%d Bonus", &m.bonuses)
	fmt.Sscanf(strings.TrimSpace(elements.Eq(5).Text()), "%d Checkpoints", &m.checkpoints)

	return m, nil
}

func main() {
	mapUrls, err := getSurfMapUrls()
	if err != nil {
		panic(err)
	}

	for _, mapUrl := range mapUrls {
		m, err := getSurfMap(mapUrl)

		if err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", m)
	}
}
