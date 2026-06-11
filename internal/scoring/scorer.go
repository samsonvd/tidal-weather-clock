package scoring

import (
	"sort"
	"time"

	"github.com/samson/tidal-weather-clock/internal/domain"
)

// ScoreDay scores all activities against the provided hourly data and returns
// a flat list of windows sorted descending by score. Excluded windows (failed
// required constraints) are included at the end for debug purposes.
func ScoreDay(activities []domain.Activity, hours []domain.HourlyData) []ScoredWindow {
	var results []ScoredWindow
	for _, activity := range activities {
		for _, window := range findWindows(activity, hours) {
			results = append(results, scoreWindow(activity, window))
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Excluded != results[j].Excluded {
			return !results[i].Excluded
		}
		return results[i].Score > results[j].Score
	})
	return results
}

// findWindows returns all slices of DurationHrs consecutive hours that fall
// within the activity's [WindowStart, WindowEnd] range.
func findWindows(a domain.Activity, hours []domain.HourlyData) [][]domain.HourlyData {
	var inWindow []domain.HourlyData
	for _, h := range hours {
		hod := h.Time.UTC().Hour()
		if hod >= a.WindowStart && hod <= a.WindowEnd {
			inWindow = append(inWindow, h)
		}
	}

	var windows [][]domain.HourlyData
	for i := 0; i <= len(inWindow)-a.DurationHrs; i++ {
		w := inWindow[i : i+a.DurationHrs]
		if consecutive(w) {
			windows = append(windows, w)
		}
	}
	return windows
}

func consecutive(hours []domain.HourlyData) bool {
	for i := 1; i < len(hours); i++ {
		if hours[i].Time.Sub(hours[i-1].Time) != time.Hour {
			return false
		}
	}
	return true
}

func scoreWindowHourScores(hourScores []float64, c domain.Constraint) ConstraintResult {
	if len(hourScores) == 0 {
		return ConstraintResult{
			Constraint: c,
			HourScores: hourScores,
			Score:      0,
			Passed:     !c.Required,
		}
	}

	var windowScore float64
	if c.Required {
		// Required: worst hour determines the window score.
		windowScore = hourScores[0]
		for _, s := range hourScores[1:] {
			if s < windowScore {
				windowScore = s
			}
		}
	} else {
		// Preferred: average across hours so one bad hour doesn't zero the score.
		sum := 0.0
		for _, s := range hourScores {
			sum += s
		}
		windowScore = sum / float64(len(hourScores))
	}
	return ConstraintResult{
		Constraint: c,
		HourScores: hourScores,
		Score:      windowScore,
		Passed:     !c.Required || windowScore > 0,
	}
}

func scoreWindow(a domain.Activity, hours []domain.HourlyData) ScoredWindow {
	results := make([]ConstraintResult, len(a.Constraints))

	for ci, c := range a.Constraints {
		hourScores := make([]float64, len(hours))
		for hi, h := range hours {
			hourScores[hi] = scoreHour(c, h)
		}
		results[ci] = scoreWindowHourScores(hourScores, c)
	}

	// Check required constraints.
	for _, cr := range results {
		if !cr.Passed {
			return ScoredWindow{
				Start:             hours[0].Time,
				End:               hours[len(hours)-1].Time.Add(time.Hour),
				Score:             0,
				Excluded:          true,
				Activity:          a,
				Hours:             hours,
				ConstraintResults: results,
			}
		}
	}

	// Weighted average of preferred constraint scores.
	var totalWeight, weightedSum float64
	for _, cr := range results {
		totalWeight += cr.Constraint.Weight
		weightedSum += cr.Score * cr.Constraint.Weight
	}

	var score float64
	if totalWeight > 0 {
		score = weightedSum / totalWeight
	}

	return ScoredWindow{
		Start:             hours[0].Time,
		End:               hours[len(hours)-1].Time.Add(time.Hour),
		Score:             score,
		Excluded:          false,
		Activity:          a,
		Hours:             hours,
		ConstraintResults: results,
	}
}
