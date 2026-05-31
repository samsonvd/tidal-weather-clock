package templates

import (
	"fmt"
	"strings"
	"time"

	_ "time/tzdata"

	"github.com/samson/tidal-weather-clock/internal/domain"
	"github.com/samson/tidal-weather-clock/internal/scoring"
)

var londonTZ *time.Location

func init() {
	var err error
	londonTZ, err = time.LoadLocation("Europe/London")
	if err != nil {
		londonTZ = time.UTC
	}
}

type DayViewData struct {
	Date     time.Time
	PrevDate string
	NextDate string
	Hours    []domain.HourlyData
	Windows  []scoring.ScoredWindow
	User     *domain.User
	HasData  bool
}

func formatTime(t time.Time) string {
	return t.In(londonTZ).Format("15:04")
}

func formatDate(t time.Time) string {
	return t.In(londonTZ).Format("Mon 2 Jan 2006")
}

func compassPoint(deg float64) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	return dirs[int((deg+22.5)/45)%8]
}

func weatherLabel(code domain.WeatherCode) string {
	labels := map[domain.WeatherCode]string{
		domain.WeatherClear:        "Clear",
		domain.WeatherPartlyCloudy: "Partly cloudy",
		domain.WeatherOvercast:     "Overcast",
		domain.WeatherFog:          "Fog",
		domain.WeatherDrizzle:      "Drizzle",
		domain.WeatherRain:         "Rain",
		domain.WeatherHeavyRain:    "Heavy rain",
		domain.WeatherSnow:         "Snow",
		domain.WeatherThunderstorm: "Thunderstorm",
	}
	if l, ok := labels[code]; ok {
		return l
	}
	return string(code)
}

func scorePercent(s float64) string {
	return fmt.Sprintf("%.0f%%", s*100)
}

func formatHourScores(scores []float64) string {
	parts := make([]string, len(scores))
	for i, s := range scores {
		parts[i] = fmt.Sprintf("%.0f%%", s*100)
	}
	return strings.Join(parts, ", ")
}
