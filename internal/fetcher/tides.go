package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"
)

type HourlyTide struct {
	Time        time.Time
	TideHeightM float64
}

type tidalEvent struct {
	DateTime string  `json:"dateTime"`
	Height   float64 `json:"height"`
}

type easyTideResponse struct {
	TidalEventList []tidalEvent `json:"tidalEventList"`
}

func FetchTides(ctx context.Context, stationID string) ([]HourlyTide, error) {
	url := fmt.Sprintf("https://easytide.admiralty.co.uk/Home/GetPredictionData?stationId=%s", stationID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("easytide returned %d", resp.StatusCode)
	}

	var body easyTideResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return interpolateTides(body.TidalEventList)
}

type turningPoint struct {
	time   time.Time
	height float64
}

func interpolateTides(events []tidalEvent) ([]HourlyTide, error) {
	if len(events) < 2 {
		return nil, fmt.Errorf("need at least 2 tidal events to interpolate")
	}

	// Parse HW events.
	hwEvents := make([]turningPoint, 0, len(events))
	for _, e := range events {
		t, err := parseEasyTideTime(e.DateTime)
		if err != nil {
			return nil, fmt.Errorf("parse tidal event time: %w", err)
		}
		hwEvents = append(hwEvents, turningPoint{time: t, height: e.Height})
	}
	sort.Slice(hwEvents, func(i, j int) bool {
		return hwEvents[i].time.Before(hwEvents[j].time)
	})

	// Build full turning point sequence: HW, synthetic LW (midpoint, 0m), HW, ...
	turning := make([]turningPoint, 0, len(hwEvents)*2-1)
	for i, hw := range hwEvents {
		turning = append(turning, hw)
		if i < len(hwEvents)-1 {
			next := hwEvents[i+1]
			mid := hw.time.Add(next.time.Sub(hw.time) / 2)
			turning = append(turning, turningPoint{time: mid, height: 0})
		}
	}

	// Determine the hourly window: first HW to last HW.
	start := turning[0].time.Truncate(time.Hour)
	end := turning[len(turning)-1].time.Truncate(time.Hour).Add(time.Hour)

	var results []HourlyTide
	for t := start; !t.After(end); t = t.Add(time.Hour) {
		h := interpolateHeight(turning, t)
		results = append(results, HourlyTide{Time: t.UTC(), TideHeightM: h})
	}
	return results, nil
}

// interpolateHeight uses cosine interpolation between the two turning points
// surrounding t. If t is outside the range it clamps to the nearest pair.
func interpolateHeight(points []turningPoint, t time.Time) float64 {
	if t.Before(points[0].time) || t.Equal(points[0].time) {
		return points[0].height
	}
	if !t.Before(points[len(points)-1].time) {
		return points[len(points)-1].height
	}

	// Find surrounding pair.
	idx := sort.Search(len(points), func(i int) bool {
		return !points[i].time.Before(t)
	})
	if idx == 0 {
		return points[0].height
	}

	a := points[idx-1]
	b := points[idx]

	span := b.time.Sub(a.time).Seconds()
	elapsed := t.Sub(a.time).Seconds()
	frac := elapsed / span

	// h(t) = (ha + hb)/2 + (ha - hb)/2 * cos(π * frac)
	return (a.height+b.height)/2 + (a.height-b.height)/2*math.Cos(math.Pi*frac)
}

func parseEasyTideTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised time format: %q", s)
}
