package templates

import (
	"fmt"
	"strings"
	"time"

	_ "time/tzdata"

	"github.com/samson/tidal-weather-clock/internal/domain"
	"github.com/samson/tidal-weather-clock/internal/scoring"
)

// formatConstraintConfig returns a human-readable description of what a
// constraint is requesting, for use in the debug panel.
func formatConstraintConfig(c domain.Constraint) string {
	switch c.Type {
	case domain.ConstraintWindSpeed:
		return fmt.Sprintf(
			"Acceptable %.0f–%.0f kts (F%d–F%d) · Ideal %.0f–%.0f kts (F%d–F%d)",
			msToKts(c.AcceptableMin), msToKts(c.AcceptableMax),
			msToBeaufort(c.AcceptableMin), msToBeaufort(c.AcceptableMax),
			msToKts(c.IdealMin), msToKts(c.IdealMax),
			msToBeaufort(c.IdealMin), msToBeaufort(c.IdealMax),
		)
	case domain.ConstraintTideHeight:
		return fmt.Sprintf(
			"Acceptable %.2f–%.2fm · Ideal %.2f–%.2fm",
			c.AcceptableMin, c.AcceptableMax, c.IdealMin, c.IdealMax,
		)
	case domain.ConstraintWindDir:
		return fmt.Sprintf("Preferred: %s (±%.0f°)", strings.Join(degsToPoints(c.Preferred), ", "), c.Tolerance)
	case domain.ConstraintWeather:
		return fmt.Sprintf("Ideal: %s · Acceptable: %s",
			strings.Join(c.Ideal, ", "), strings.Join(c.Acceptable, ", "))
	}
	return ""
}

// formatConstraintValue returns the relevant measured value from an hourly
// reading for the given constraint type.
func formatConstraintValue(cType domain.ConstraintType, h domain.HourlyData) string {
	switch cType {
	case domain.ConstraintWindSpeed:
		return fmt.Sprintf("%.0f kts / F%d", msToKts(h.WindSpeedMS), msToBeaufort(h.WindSpeedMS))
	case domain.ConstraintTideHeight:
		return fmt.Sprintf("%.2fm", h.TideHeightM)
	case domain.ConstraintWindDir:
		return fmt.Sprintf("%s (%.0f°)", compassPoint(h.WindDirDeg), h.WindDirDeg)
	case domain.ConstraintWeather:
		return weatherLabel(h.Weather)
	}
	return ""
}

// Ensure scoring is imported (used by DayViewData).
var _ = scoring.ScoredWindow{}

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

func formatBool(b bool, trueStr, falseStr string) string {
	if b {
		return trueStr
	}
	return falseStr
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

const ktsPerMS = 1.0 / 0.514444

func msToKts(ms float64) float64  { return ms * ktsPerMS }
func ktsToMS(kts float64) float64 { return kts * 0.514444 }

func msToBeaufort(ms float64) int {
	switch {
	case ms < 0.3:
		return 0
	case ms < 1.6:
		return 1
	case ms < 3.4:
		return 2
	case ms < 5.5:
		return 3
	case ms < 8.0:
		return 4
	case ms < 10.8:
		return 5
	case ms < 13.9:
		return 6
	case ms < 17.2:
		return 7
	case ms < 20.8:
		return 8
	case ms < 24.5:
		return 9
	case ms < 28.5:
		return 10
	case ms < 32.7:
		return 11
	default:
		return 12
	}
}

func formatWind(ms float64, dirDeg float64) string {
	return fmt.Sprintf("%.0f kts / F%d %s", msToKts(ms), msToBeaufort(ms), compassPoint(dirDeg))
}

func formatHour(h int) string {
	return fmt.Sprintf("%02d:00", h)
}

type CompassPoint struct {
	Label string
	Deg   float64
}

var CompassPoints = []CompassPoint{
	{"N", 0}, {"NE", 45}, {"E", 90}, {"SE", 135},
	{"S", 180}, {"SW", 225}, {"W", 270}, {"NW", 315},
}

func degsToPoints(dirs []float64) []string {
	degToPoint := map[float64]string{
		0: "N", 45: "NE", 90: "E", 135: "SE",
		180: "S", 225: "SW", 270: "W", 315: "NW",
	}
	points := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if p, ok := degToPoint[d]; ok {
			points = append(points, p)
		}
	}
	return points
}

type WeatherOption struct {
	Code  domain.WeatherCode
	Label string
}

var WeatherOptions = []WeatherOption{
	{domain.WeatherClear, "Clear"},
	{domain.WeatherPartlyCloudy, "Partly cloudy"},
	{domain.WeatherOvercast, "Overcast"},
	{domain.WeatherFog, "Fog"},
	{domain.WeatherDrizzle, "Drizzle"},
	{domain.WeatherRain, "Rain"},
	{domain.WeatherHeavyRain, "Heavy rain"},
	{domain.WeatherSnow, "Snow"},
	{domain.WeatherThunderstorm, "Thunderstorm"},
}

func sliceContains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func formatHourScores(scores []float64) string {
	parts := make([]string, len(scores))
	for i, s := range scores {
		parts[i] = fmt.Sprintf("%.0f%%", s*100)
	}
	return strings.Join(parts, ", ")
}
