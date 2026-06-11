package scoring

import (
	"math"

	"github.com/samson/tidal-weather-clock/internal/domain"
)

func scoreHour(c domain.Constraint, h domain.HourlyData) float64 {
	switch c.Type {
	case domain.ConstraintWindSpeed:
		return trapezoid(h.WindSpeedMS, c.AcceptableMin, c.IdealMin, c.IdealMax, c.AcceptableMax)
	case domain.ConstraintTideHeight:
		return trapezoid(h.TideHeightM, c.AcceptableMin, c.IdealMin, c.IdealMax, c.AcceptableMax)
	case domain.ConstraintWindDir:
		return scoreDirection(h.WindDirDeg, c.Preferred, c.Tolerance)
	case domain.ConstraintWeather:
		return scoreWeather(h.Weather, c.Acceptable, c.Ideal)
	}
	return 0
}

// trapezoid returns 1.0 inside [idealMin, idealMax], falls off linearly to 0.0
// at the edges of [acceptableMin, acceptableMax], and 0.0 outside.
func trapezoid(val, acceptMin, idealMin, idealMax, acceptMax float64) float64 {
	if val < acceptMin || val > acceptMax {
		return 0
	}
	if val >= idealMin && val <= idealMax {
		return 1
	}
	if val < idealMin {
		return (val - acceptMin) / (idealMin - acceptMin)
	}
	return (acceptMax - val) / (acceptMax - idealMax)
}

// scoreDirection returns 1.0 if dir is within tolerance degrees of any preferred
// direction, accounting for the 0/360 wraparound, and 0.0 otherwise.
func scoreDirection(dir float64, preferred []float64, tolerance float64) float64 {
	for _, p := range preferred {
		diff := math.Abs(dir - p)
		if diff > 180 {
			diff = 360 - diff
		}
		if diff <= tolerance {
			return 1
		}
	}
	return 0
}

// scoreWeather returns 1.0 if the code is in ideal, 0.5 if only in acceptable,
// and 0.0 otherwise.
func scoreWeather(code domain.WeatherCode, acceptable, ideal []string) float64 {
	for _, i := range ideal {
		if string(code) == i {
			return 1
		}
	}
	for _, a := range acceptable {
		if string(code) == a {
			return 0.5
		}
	}
	return 0
}
