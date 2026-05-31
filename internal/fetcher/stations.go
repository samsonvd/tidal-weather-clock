package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type Station struct {
	ID      string
	Name    string
	Country string
	Lat     float64
	Lon     float64
}

type StationGroup struct {
	Country  string
	Stations []Station
}

type stationsResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` // [lon, lat]
		} `json:"geometry"`
		Properties struct {
			ID      string `json:"Id"`
			Name    string `json:"Name"`
			Country string `json:"Country"`
		} `json:"properties"`
	} `json:"features"`
}

func FetchStations(ctx context.Context) ([]Station, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://easytide.admiralty.co.uk/Home/GetStations", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("easytide stations returned %d", resp.StatusCode)
	}

	var body stationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	stations := make([]Station, 0, len(body.Features))
	for _, f := range body.Features {
		if len(f.Geometry.Coordinates) < 2 {
			continue
		}
		stations = append(stations, Station{
			ID:      f.Properties.ID,
			Name:    f.Properties.Name,
			Country: f.Properties.Country,
			Lon:     f.Geometry.Coordinates[0],
			Lat:     f.Geometry.Coordinates[1],
		})
	}
	return stations, nil
}

func GroupByCountry(stations []Station) []StationGroup {
	index := make(map[string][]Station)
	order := []string{}
	for _, s := range stations {
		if _, ok := index[s.Country]; !ok {
			order = append(order, s.Country)
		}
		index[s.Country] = append(index[s.Country], s)
	}
	sort.Strings(order)

	groups := make([]StationGroup, 0, len(order))
	for _, country := range order {
		ss := index[country]
		sort.Slice(ss, func(i, j int) bool { return ss[i].Name < ss[j].Name })
		groups = append(groups, StationGroup{Country: country, Stations: ss})
	}
	return groups
}
