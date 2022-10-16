package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const SURF_MAPS_URL = "https://surfheaven.eu/player/43223876"
const SURFHEAVEN_URL = "https://surfheaven.eu"

type surfMap struct {
	name        string
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
		return nil, errors.New("Error parsing HTML for SurfHeaven maps")
	}

	mapUrls := make([]string, 0)
	doc.Find(".table-maps .table tbody tr td a").Each(func(i int, s *goquery.Selection) {
		mapHref, ok := s.Attr("href")
		if ok {
			mapUrls = append(mapUrls, mapHref)
		}
	})

	sort.Strings(mapUrls)

	return mapUrls, nil
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
		return surfMap{}, errors.New("Error parsing HTML for SurfHeaven map url")
	}

	elements := doc.Find(".table.table-responsive.nodatatable tbody tr td")

	var m surfMap

	m.name = mapUrl[len("/map/"):]
	fmt.Sscanf(strings.TrimSpace(elements.Eq(0).Text()), "%d Completions", &m.completions)
	fmt.Sscanf(strings.TrimSpace(elements.Eq(1).Text()), "%d Times Played", &m.timesPlayed)
	fmt.Sscanf(strings.TrimSpace(elements.Eq(2).Text()), "%d Tier", &m.tier)
	m.kind = strings.TrimSpace(elements.Eq(3).Text())
	fmt.Sscanf(strings.TrimSpace(elements.Eq(4).Text()), "%d Bonus", &m.bonuses)
	fmt.Sscanf(strings.TrimSpace(elements.Eq(5).Text()), "%d Checkpoints", &m.checkpoints)

	return m, nil
}

func surfMapToSlice(m surfMap) []string {
	return []string{
		m.name,
		strconv.Itoa(m.completions),
		strconv.Itoa(m.timesPlayed),
		strconv.Itoa(m.tier),
		m.kind,
		strconv.Itoa(m.bonuses),
		strconv.Itoa(m.checkpoints),
	}
}

func main() {
	mapUrls, err := getSurfMapUrls()
	if err != nil {
		panic(err)
	}

	csvFile, err := os.Create("maps.csv")
	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)

	writer.Write([]string{"name", "completions", "timesPlayed", "tier", "kind", "bonuses", "checkpoints"})

	for idx, mapUrl := range mapUrls {
		time.Sleep(time.Second / 10)
		m, err := getSurfMap(mapUrl)

		if err != nil {
			panic(err)
		}

		fmt.Printf("[%3d/%3d] Writing %s\n", idx, len(mapUrls), m.name)
		writer.Write(surfMapToSlice(m))
		writer.Flush()
	}
}
